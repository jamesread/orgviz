package orgs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const sampleTheme = `
title: Ocean
backgroundColor: "#f2f2f2"
features:
  team: true
  expandButtons: false
css: |
  div.person { border-top-color: teal; }
`

func TestParseThemeReadsFields(t *testing.T) {
	theme, err := ParseTheme([]byte(sampleTheme))
	require.NoError(t, err)
	require.Equal(t, "Ocean", theme.Title)
	require.Equal(t, "#f2f2f2", theme.Style.BackgroundColor)
	require.Contains(t, theme.Style.CSS, "border-top-color")
	require.True(t, theme.Style.Features.Resolve().Team)
	require.False(t, theme.Style.Features.Resolve().ExpandButtons)
}

func TestLoadThemesUsesFileNameAsID(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ocean.yml"), []byte(sampleTheme), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("ignore"), 0o644))

	loaded, err := LoadThemes(dir)
	require.NoError(t, err)
	require.Len(t, loaded, 1)
	require.Equal(t, "ocean", loaded[0].ID)
	require.Equal(t, "Ocean", loaded[0].Title)
}

func TestLoadThemesEmptyDir(t *testing.T) {
	loaded, err := LoadThemes("")
	require.NoError(t, err)
	require.Empty(t, loaded)

	loaded, err = LoadThemes(t.TempDir() + "/missing")
	require.NoError(t, err)
	require.Empty(t, loaded)
}

func TestStyleFeaturesCompactDefault(t *testing.T) {
	require.True(t, StyleFeatures{}.Resolve().Compact)

	off := false
	require.False(t, StyleFeatures{Compact: &off}.Resolve().Compact)
}

func TestStyleOverlayPrefersChartFields(t *testing.T) {
	team := true
	chart := Style{Theme: "ocean", CSS: "h3 { color: navy; }", Features: StyleFeatures{Team: &team}}
	base := Style{CSS: "div.person { color: red; }", BackgroundColor: "#eee", Features: StyleFeatures{}}

	merged := chart.Overlay(base)
	require.Equal(t, "ocean", merged.Theme)
	require.Equal(t, "h3 { color: navy; }", merged.CSS)
	require.Equal(t, "#eee", merged.BackgroundColor)
	require.True(t, merged.Features.Resolve().Team)
}

func TestValidThemeID(t *testing.T) {
	require.True(t, ValidThemeID("ocean"))
	require.False(t, ValidThemeID("../x"))
}
