package charts

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jamesread/orgviz/internal/orgs"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
)

// ListThemes returns the shared themes in ID order.
func (s *Store) ListThemes() []*orgs.Theme {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return append([]*orgs.Theme(nil), s.themes...)
}

// GetTheme finds a shared theme by ID.
func (s *Store) GetTheme(id string) (*orgs.Theme, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.themeByID(id)
}

func (s *Store) themeByID(id string) (*orgs.Theme, bool) {
	for _, theme := range s.themes {
		if theme.ID == id {
			return theme, true
		}
	}

	return nil, false
}

// ResolveStyle applies a named theme under the chart's own style fields.
func (s *Store) ResolveStyle(style orgs.Style) orgs.Style {
	if style.Theme == "" {
		return style
	}

	theme, ok := s.GetTheme(style.Theme)
	if !ok {
		return style
	}

	return style.Overlay(theme.Style)
}

// ThemeWarning reports a missing style.theme target.
func (s *Store) ThemeWarning(style orgs.Style) string {
	if style.Theme == "" {
		return ""
	}

	if _, ok := s.GetTheme(style.Theme); ok {
		return ""
	}

	return fmt.Sprintf("theme %q is not saved yet", style.Theme)
}

// SaveTheme writes a shared theme, replacing one with the same ID.
func (s *Store) SaveTheme(theme *orgs.Theme) (*orgs.Theme, error) {
	if err := checkTheme(theme); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.writeThemeFile(theme); err != nil {
		return nil, err
	}

	s.putTheme(theme)
	log.WithField("theme", theme.ID).Info("Saved theme file")

	return theme, nil
}

func (s *Store) writeThemeFile(theme *orgs.Theme) error {
	if s.themesDir == "" {
		return ErrReadOnly
	}

	if err := os.MkdirAll(s.themesDir, 0o755); err != nil {
		return fmt.Errorf("creating themes directory: %w", err)
	}

	source, err := marshalTheme(theme)
	if err != nil {
		return err
	}

	return writeAtomic(s.themePath(theme.ID), source)
}

func checkTheme(theme *orgs.Theme) error {
	if theme == nil {
		return fmt.Errorf("%w theme id: use letters, digits, '-' and '_' (up to 64)", ErrInvalid)
	}

	if !orgs.ValidThemeID(theme.ID) {
		return fmt.Errorf("%w theme id %q: use letters, digits, '-' and '_' (up to 64)", ErrInvalid, theme.ID)
	}

	if theme.Title == "" {
		theme.Title = theme.ID
	}

	return nil
}

func marshalTheme(theme *orgs.Theme) ([]byte, error) {
	data, err := yaml.Marshal(theme)
	if err != nil {
		return nil, fmt.Errorf("%w: writing theme: %v", ErrInvalid, err)
	}

	return data, nil
}

func (s *Store) putTheme(theme *orgs.Theme) {
	for i, existing := range s.themes {
		if existing.ID == theme.ID {
			s.themes[i] = theme
			return
		}
	}

	s.themes = append(s.themes, theme)
	sort.SliceStable(s.themes, func(a, b int) bool {
		return s.themes[a].ID < s.themes[b].ID
	})
}

func (s *Store) themePath(id string) string {
	return filepath.Join(s.themesDir, id+".yml")
}
