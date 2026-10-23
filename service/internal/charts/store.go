// Package charts holds the loaded org charts and writes changes to them back
// to the orgs and avatars directories.
package charts

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"

	"github.com/jamesread/orgviz/internal/orgs"
	log "github.com/sirupsen/logrus"
)

// MaxSourceBytes caps the size of an org file accepted from a client.
const MaxSourceBytes = 1 << 20

var (
	ErrNotFound = errors.New("chart not found")
	ErrExists   = errors.New("already exists")
	ErrInvalid  = errors.New("invalid")
	ErrReadOnly = errors.New("no directory to write to")
)

// Store is safe for concurrent use.
type Store struct {
	mu         sync.RWMutex
	orgsDir    string
	avatarsDir string
	themesDir  string
	charts     []*orgs.File
	themes     []*orgs.Theme
}

// New loads every org file in orgsDir and theme in themesDir. A directory
// may be empty, which makes the matching writes fail with ErrReadOnly.
func New(orgsDir, avatarsDir, themesDir string) *Store {
	store := &Store{orgsDir: orgsDir, avatarsDir: avatarsDir, themesDir: themesDir}

	if orgsDir != "" {
		loaded, err := orgs.LoadDir(orgsDir)
		if err != nil {
			log.WithError(err).Error("Failed to load org files")
		}
		store.charts = loaded
	}

	store.themes = loadThemes(themesDir)
	store.sort()

	return store
}

func loadThemes(dir string) []*orgs.Theme {
	loaded, err := orgs.LoadThemes(dir)
	if err != nil {
		log.WithError(err).Error("Failed to load theme files")
		return nil
	}

	return loaded
}

// NewFromFiles serves charts already in memory, with nowhere to write.
func NewFromFiles(files ...*orgs.File) *Store {
	return &Store{charts: files}
}

// List returns the charts in ID order.
func (s *Store) List() []*orgs.File {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return append([]*orgs.File(nil), s.charts...)
}

// Get finds a chart by ID, or by its position in List for old numeric links.
func (s *Store) Get(id string) (*orgs.File, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	i := s.indexOf(id)
	if i < 0 {
		i = s.legacyIndex(id)
	}

	if i < 0 {
		return nil, false
	}

	return s.charts[i], true
}

// legacyIndex reads the numeric chart IDs used before charts were named after
// their files. It returns -1 when id is not a valid position.
func (s *Store) legacyIndex(id string) int {
	i, err := strconv.Atoi(id)
	if err != nil || i < 0 || i >= len(s.charts) {
		return -1
	}

	return i
}

// Create writes a new org file. It fails if the chart already exists.
func (s *Store) Create(id, source string) (*orgs.File, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.checkNew(id); err != nil {
		return nil, err
	}

	return s.write(id, []byte(source))
}

// Update replaces the YAML of an existing org file.
func (s *Store) Update(id, source string) (*orgs.File, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.indexOf(id) < 0 {
		return nil, fmt.Errorf("%w: %q", ErrNotFound, id)
	}

	return s.write(id, []byte(source))
}

// SetStyle changes the style block of an existing org file.
func (s *Store) SetStyle(id string, change orgs.StyleChange) (*orgs.File, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	i := s.indexOf(id)
	if i < 0 {
		return nil, fmt.Errorf("%w: %q", ErrNotFound, id)
	}

	source, err := orgs.SetStyle([]byte(s.charts[i].Source), change)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalid, err)
	}

	return s.write(id, source)
}

func (s *Store) checkNew(id string) error {
	if !orgs.ValidChartID(id) {
		return fmt.Errorf("%w chart id %q: use letters, digits, '-' and '_' (up to 64)", ErrInvalid, id)
	}

	if s.indexOf(id) >= 0 || fileExists(s.chartPath(id)) {
		return fmt.Errorf("chart %q %w", id, ErrExists)
	}

	return nil
}

// write validates source, saves it as <id>.yml and swaps it into memory. The
// caller holds the write lock.
func (s *Store) write(id string, source []byte) (*orgs.File, error) {
	chart, err := parseSource(source)
	if err != nil {
		return nil, err
	}

	if s.orgsDir == "" {
		return nil, ErrReadOnly
	}

	if err := writeAtomic(s.chartPath(id), source); err != nil {
		return nil, err
	}

	chart.ID = id
	s.put(chart)
	log.WithFields(log.Fields{"chart": id, "people": len(chart.People)}).Info("Saved org file")

	return chart, nil
}

func parseSource(source []byte) (*orgs.File, error) {
	if len(source) > MaxSourceBytes {
		return nil, fmt.Errorf("%w: org YAML is larger than %d bytes", ErrInvalid, MaxSourceBytes)
	}

	chart, err := orgs.Parse(source)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalid, err)
	}

	return chart, nil
}

func (s *Store) put(chart *orgs.File) {
	if i := s.indexOf(chart.ID); i >= 0 {
		s.charts[i] = chart
		return
	}

	s.charts = append(s.charts, chart)
	s.sort()
}

func (s *Store) indexOf(id string) int {
	for i, chart := range s.charts {
		if chart.ID == id {
			return i
		}
	}

	return -1
}

func (s *Store) sort() {
	sort.SliceStable(s.charts, func(a, b int) bool {
		return s.charts[a].ID < s.charts[b].ID
	})
}

func (s *Store) chartPath(id string) string {
	return filepath.Join(s.orgsDir, id+".yml")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// writeAtomic writes through a temporary file, so readers never see half a file.
func writeAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".orgviz-*")
	if err != nil {
		return fmt.Errorf("saving %s: %w", filepath.Base(path), err)
	}
	defer os.Remove(tmp.Name())

	if err := writeAndClose(tmp, data); err != nil {
		return fmt.Errorf("saving %s: %w", filepath.Base(path), err)
	}

	return os.Rename(tmp.Name(), path)
}

func writeAndClose(file *os.File, data []byte) error {
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}

	if err := file.Chmod(0o644); err != nil {
		file.Close()
		return err
	}

	return file.Close()
}
