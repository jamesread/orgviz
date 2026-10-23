package charts

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/jamesread/orgviz/internal/orgs"
)

// MaxAvatarBytes caps the size of an uploaded avatar image.
const MaxAvatarBytes = 2 << 20

var avatarNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}\.(png|jpg|jpeg|webp|svg)$`)

var avatarTypes = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".webp": "image/webp",
	".svg":  "image/svg+xml",
}

// SaveAvatar stores an image in the avatars directory and returns its URL.
// The content must match the file extension.
func (s *Store) SaveAvatar(name string, data []byte, overwrite bool) (string, error) {
	if err := checkAvatar(name, data); err != nil {
		return "", err
	}

	path, err := s.avatarPath(name, overwrite)
	if err != nil {
		return "", err
	}

	if err := writeAtomic(path, data); err != nil {
		return "", err
	}

	return "/avatars/" + name, nil
}

func (s *Store) avatarPath(name string, overwrite bool) (string, error) {
	if s.avatarsDir == "" {
		return "", ErrReadOnly
	}

	path := filepath.Join(s.avatarsDir, name)
	if !overwrite && fileExists(path) {
		return "", fmt.Errorf("avatar %q %w; pass overwrite to replace it", name, ErrExists)
	}

	return path, nil
}

func checkAvatar(name string, data []byte) error {
	if !avatarNamePattern.MatchString(name) {
		return fmt.Errorf("%w avatar name %q: use lowercase letters, digits, '-' and '_', ending .png, .jpg, .jpeg, .webp or .svg", ErrInvalid, name)
	}

	if len(data) > MaxAvatarBytes {
		return fmt.Errorf("%w: avatar is larger than %d bytes", ErrInvalid, MaxAvatarBytes)
	}

	return checkAvatarType(name, data)
}

func checkAvatarType(name string, data []byte) error {
	want := avatarTypes[filepath.Ext(name)]
	if got := detectAvatarType(data); got != want {
		return fmt.Errorf("%w: avatar %q contains %s, not %s", ErrInvalid, name, got, want)
	}

	return nil
}

func detectAvatarType(data []byte) string {
	if isSVG(data) {
		return "image/svg+xml"
	}

	return http.DetectContentType(data)
}

func isSVG(data []byte) bool {
	start := bytes.ToLower(bytes.TrimSpace(bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))))
	if len(start) == 0 || bytes.Contains(start, []byte("<script")) {
		return false
	}

	return svgStarts(start)
}

func svgStarts(start []byte) bool {
	if bytes.HasPrefix(start, []byte("<svg")) {
		return true
	}

	head := start
	if len(head) > 1024 {
		head = head[:1024]
	}

	return bytes.HasPrefix(start, []byte("<?xml")) && bytes.Contains(head, []byte("<svg"))
}

// ListAvatars returns the image file names in the avatars directory.
func (s *Store) ListAvatars() ([]string, error) {
	if s.avatarsDir == "" {
		return nil, ErrReadOnly
	}

	entries, err := os.ReadDir(s.avatarsDir)
	if err != nil {
		return nil, fmt.Errorf("listing avatars: %w", err)
	}

	return imageNames(entries), nil
}

func imageNames(entries []os.DirEntry) []string {
	names := []string{}
	for _, entry := range entries {
		if _, ok := avatarTypes[filepath.Ext(entry.Name())]; ok && !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	return names
}

// MissingAvatars lists the `avatar` files a chart names that are not in the
// avatars directory.
func (s *Store) MissingAvatars(chart *orgs.File) []string {
	missing := []string{}

	for _, person := range chart.People {
		if person.Avatar != "" && !fileExists(filepath.Join(s.avatarsDir, filepath.Base(person.Avatar))) {
			missing = append(missing, person.Avatar)
		}
	}

	return missing
}
