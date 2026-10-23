package orgs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// Theme is a shared style saved as themes/<id>.yml.
type Theme struct {
	ID    string `yaml:"-"`
	Title string
	Style Style `yaml:",inline"`
}

// ParseTheme unmarshals one theme YAML document.
func ParseTheme(contents []byte) (*Theme, error) {
	theme := &Theme{}
	if err := yaml.Unmarshal(contents, theme); err != nil {
		return nil, fmt.Errorf("parsing theme file: %w", err)
	}

	return theme, nil
}

// LoadThemes reads every *.yml file in dir, in lexical order.
func LoadThemes(dir string) ([]*Theme, error) {
	matches, err := themeFiles(dir)
	if err != nil {
		return nil, err
	}

	themes := make([]*Theme, 0, len(matches))
	for _, filename := range matches {
		theme, err := loadThemeFile(filename)
		if err != nil {
			return nil, err
		}
		themes = append(themes, theme)
	}

	return themes, nil
}

func themeFiles(dir string) ([]string, error) {
	if dir == "" {
		return nil, nil
	}

	matches, err := filepath.Glob(filepath.Join(dir, "*.yml"))
	if err != nil {
		return nil, fmt.Errorf("listing theme files: %w", err)
	}

	return matches, nil
}

func loadThemeFile(filename string) (*Theme, error) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading theme file %s: %w", filename, err)
	}

	theme, err := ParseTheme(contents)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}

	theme.ID = strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	if theme.Title == "" {
		theme.Title = theme.ID
	}

	return theme, nil
}

// ValidThemeID reports whether id is safe to use as a theme file name.
func ValidThemeID(id string) bool {
	return ValidChartID(id)
}
