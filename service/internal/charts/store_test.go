package charts

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesread/orgviz/internal/orgs"
	"github.com/stretchr/testify/require"
)

const acme = "title: Acme\npeople:\n  - name: Alice\n"

// A 1x1 PNG.
var pngBytes = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15\xc4\x89\x00\x00\x00\rIDATx\x9cc\xf8\x0f\x00\x00\x01\x01\x00\x05\x18\xd8N\x00\x00\x00\x00IEND\xaeB`\x82")

func newTestStore(t *testing.T) (*Store, string, string) {
	t.Helper()
	orgsDir := t.TempDir()
	avatarsDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(orgsDir, "Existing.yml"), []byte(acme), 0o644))

	return New(orgsDir, avatarsDir, t.TempDir()), orgsDir, avatarsDir
}

func TestGetByIDOrIndex(t *testing.T) {
	store, _, _ := newTestStore(t)

	byID, ok := store.Get("Existing")
	require.True(t, ok)
	require.Equal(t, "Acme", byID.Title)

	byIndex, ok := store.Get("0")
	require.True(t, ok)
	require.Equal(t, byID, byIndex)

	_, ok = store.Get("Missing")
	require.False(t, ok)
}

func TestCreateWritesFileAndServesIt(t *testing.T) {
	store, orgsDir, _ := newTestStore(t)

	chart, err := store.Create("NewCo", "title: NewCo\npeople:\n  - name: Bob\n")
	require.NoError(t, err)
	require.Equal(t, "NewCo", chart.ID)

	written, err := os.ReadFile(filepath.Join(orgsDir, "NewCo.yml"))
	require.NoError(t, err)
	require.Contains(t, string(written), "Bob")

	got, ok := store.Get("NewCo")
	require.True(t, ok)
	require.Equal(t, "NewCo", got.Title)
	require.Len(t, store.List(), 2)
}

func TestCreateRejectsExistingInvalidIDAndBadYAML(t *testing.T) {
	store, _, _ := newTestStore(t)

	_, err := store.Create("Existing", acme)
	require.ErrorIs(t, err, ErrExists)

	_, err = store.Create("../escape", acme)
	require.ErrorIs(t, err, ErrInvalid)

	_, err = store.Create("Bad", "people: [")
	require.ErrorIs(t, err, ErrInvalid)
}

func TestUpdateReplacesYAML(t *testing.T) {
	store, orgsDir, _ := newTestStore(t)

	_, err := store.Update("Existing", "title: Renamed\npeople:\n  - name: Carol\n")
	require.NoError(t, err)

	got, _ := store.Get("Existing")
	require.Equal(t, "Renamed", got.Title)

	written, err := os.ReadFile(filepath.Join(orgsDir, "Existing.yml"))
	require.NoError(t, err)
	require.Contains(t, string(written), "Carol")

	_, err = store.Update("Missing", acme)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestUpdateKeepsOldChartWhenYAMLIsBad(t *testing.T) {
	store, _, _ := newTestStore(t)

	_, err := store.Update("Existing", "people: [")
	require.ErrorIs(t, err, ErrInvalid)

	got, _ := store.Get("Existing")
	require.Equal(t, "Acme", got.Title)
}

func TestSetStyle(t *testing.T) {
	store, _, _ := newTestStore(t)

	chart, err := store.SetStyle("Existing", orgs.StyleChange{Features: map[string]bool{"team": true}})
	require.NoError(t, err)
	require.True(t, chart.Style.Features.Resolve().Team)

	_, err = store.SetStyle("Existing", orgs.StyleChange{Features: map[string]bool{"nope": true}})
	require.ErrorIs(t, err, ErrInvalid)
}

func TestSaveAvatar(t *testing.T) {
	store, _, avatarsDir := newTestStore(t)

	url, err := store.SaveAvatar("dana.png", pngBytes, false)
	require.NoError(t, err)
	require.Equal(t, "/avatars/dana.png", url)

	written, err := os.ReadFile(filepath.Join(avatarsDir, "dana.png"))
	require.NoError(t, err)
	require.Equal(t, pngBytes, written)

	_, err = store.SaveAvatar("dana.png", pngBytes, false)
	require.ErrorIs(t, err, ErrExists)

	_, err = store.SaveAvatar("dana.png", pngBytes, true)
	require.NoError(t, err)

	names, err := store.ListAvatars()
	require.NoError(t, err)
	require.Equal(t, []string{"dana.png"}, names)
}

var svgBytes = []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="1" height="1"/>`)

func TestSaveAvatarAcceptsSVG(t *testing.T) {
	store, _, avatarsDir := newTestStore(t)

	url, err := store.SaveAvatar("dana.svg", svgBytes, false)
	require.NoError(t, err)
	require.Equal(t, "/avatars/dana.svg", url)

	written, err := os.ReadFile(filepath.Join(avatarsDir, "dana.svg"))
	require.NoError(t, err)
	require.Equal(t, svgBytes, written)

	names, err := store.ListAvatars()
	require.NoError(t, err)
	require.Equal(t, []string{"dana.svg"}, names)

	withXML := []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg" width="1" height="1"/>`)
	_, err = store.SaveAvatar("erin.svg", withXML, false)
	require.NoError(t, err)
}

func TestSaveAvatarRejectsScriptedOrMismatchedSVG(t *testing.T) {
	store, _, _ := newTestStore(t)

	_, err := store.SaveAvatar("dana.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`), false)
	require.ErrorIs(t, err, ErrInvalid)

	_, err = store.SaveAvatar("dana.svg", []byte("<html><svg></svg></html>"), false)
	require.ErrorIs(t, err, ErrInvalid)

	_, err = store.SaveAvatar("dana.svg", pngBytes, false)
	require.ErrorIs(t, err, ErrInvalid)
}

func TestSaveAvatarRejectsBadInput(t *testing.T) {
	store, _, _ := newTestStore(t)

	for _, name := range []string{"../x.png", "Dana.PNG", "dana.gif", "dana", ".png"} {
		_, err := store.SaveAvatar(name, pngBytes, false)
		require.ErrorIs(t, err, ErrInvalid, name)
	}

	_, err := store.SaveAvatar("dana.jpg", pngBytes, false)
	require.ErrorIs(t, err, ErrInvalid, "png content with a jpg name")

	_, err = store.SaveAvatar("dana.png", []byte("not an image"), false)
	require.ErrorIs(t, err, ErrInvalid)

	_, err = store.SaveAvatar("dana.png", make([]byte, MaxAvatarBytes+1), false)
	require.ErrorIs(t, err, ErrInvalid)
}

func TestWritesFailWithoutDirectories(t *testing.T) {
	store := NewFromFiles()

	_, err := store.Create("NewCo", acme)
	require.ErrorIs(t, err, ErrReadOnly)

	_, err = store.SaveAvatar("dana.png", pngBytes, false)
	require.ErrorIs(t, err, ErrReadOnly)

	_, err = store.SaveTheme(&orgs.Theme{ID: "ocean", Title: "Ocean"})
	require.ErrorIs(t, err, ErrReadOnly)
}

func TestSaveAndResolveTheme(t *testing.T) {
	store, _, _ := newTestStore(t)
	team := true
	saved, err := store.SaveTheme(&orgs.Theme{
		ID:    "ocean",
		Title: "Ocean",
		Style: orgs.Style{CSS: "h3 { color: red; }", BackgroundColor: "#eee", Features: orgs.StyleFeatures{Team: &team}},
	})
	require.NoError(t, err)
	require.Equal(t, "ocean", saved.ID)

	themes := store.ListThemes()
	require.Len(t, themes, 1)
	require.Equal(t, "Ocean", themes[0].Title)

	got, ok := store.GetTheme("ocean")
	require.True(t, ok)
	require.Equal(t, "h3 { color: red; }", got.Style.CSS)

	resolved := store.ResolveStyle(orgs.Style{Theme: "ocean", CSS: "h3 { color: navy; }"})
	require.Equal(t, "h3 { color: navy; }", resolved.CSS)
	require.Equal(t, "#eee", resolved.BackgroundColor)
	require.True(t, resolved.Features.Resolve().Team)

	require.Equal(t, `theme "missing" is not saved yet`, store.ThemeWarning(orgs.Style{Theme: "missing"}))
	require.Equal(t, "", store.ThemeWarning(orgs.Style{Theme: "ocean"}))
}

func TestSaveThemeRejectsBadID(t *testing.T) {
	store, _, _ := newTestStore(t)

	_, err := store.SaveTheme(&orgs.Theme{ID: "../x"})
	require.ErrorIs(t, err, ErrInvalid)
}

func TestMissingAvatars(t *testing.T) {
	store, _, _ := newTestStore(t)
	_, err := store.SaveAvatar("dana.png", pngBytes, false)
	require.NoError(t, err)

	chart, err := orgs.Parse([]byte("people:\n  - name: Dana\n    avatar: dana.png\n  - name: Eve\n    avatar: eve.png\n  - name: Fred\n"))
	require.NoError(t, err)

	require.Equal(t, []string{"eve.png"}, store.MissingAvatars(chart))
}

func TestExportPNG(t *testing.T) {
	store, _, _ := newTestStore(t)

	data, name, err := store.ExportPNG("Existing")
	require.NoError(t, err)
	require.Equal(t, "acme.png", name)
	require.True(t, len(data) > 8)
	require.Equal(t, []byte("\x89PNG"), data[:4])

	_, _, err = store.ExportPNG("Missing")
	require.ErrorIs(t, err, ErrNotFound)
}
