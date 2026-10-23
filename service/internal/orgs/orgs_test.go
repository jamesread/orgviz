package orgs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const sampleOrg = `
title: Example Company
people:
  - name: Alice Smith
    title: CEO
    alias: ali
  - name: Bob Johnson
    title: Product Manager
    alias: bob
    reports: ali
  - name: Lisa-Marie "Liz" Johnson
    title: Design Lead
    reports: bob
`

func TestParseAssignsStableIDsForAliases(t *testing.T) {
	org, err := Parse([]byte(sampleOrg))
	require.NoError(t, err)
	require.Equal(t, "Example Company", org.Title)
	require.Len(t, org.People, 3)

	require.Equal(t, int32(1), org.People[0].ID)
	require.Equal(t, int32(2), org.People[1].ID)
	require.Equal(t, int32(3), org.People[2].ID)

	require.Equal(t, int32(1), org.People[1].ReportsToID)
	require.Equal(t, int32(2), org.People[2].ReportsToID)
}

func TestParseRejectsInvalidYAML(t *testing.T) {
	_, err := Parse([]byte("people: ["))
	require.Error(t, err)
}

func TestLoadDirReadsYamlFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "example.yml"), []byte(sampleOrg), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("ignore"), 0o644))

	loaded, err := LoadDir(dir)
	require.NoError(t, err)
	require.Len(t, loaded, 1)
	require.Equal(t, "Example Company", loaded[0].Title)
}

func TestAvatarURLUsesFirstName(t *testing.T) {
	require.Equal(t, "/avatars/alice.jpg", AvatarURL("Alice Smith"))
	require.Equal(t, "/avatars/lisa-marie.jpg", AvatarURL("Lisa-Marie \"Liz\" Johnson"))
	require.Equal(t, "", AvatarURL(""))
}

func TestParseKeepsSource(t *testing.T) {
	org, err := Parse([]byte(sampleOrg))
	require.NoError(t, err)
	require.Equal(t, sampleOrg, org.Source)
}

func TestStyleDefaultsWhenMissing(t *testing.T) {
	org, err := Parse([]byte(sampleOrg))
	require.NoError(t, err)

	features := org.Style.Features.Resolve()
	require.Equal(t, "", org.Style.CSS)
	require.Equal(t, "", org.Style.Theme)
	require.Equal(t, "", org.Style.BackgroundColor)
	require.True(t, features.Avatars)
	require.True(t, features.JobTitles)
	require.False(t, features.Team)
	require.True(t, features.ExpandButtons)
}

func TestStyleReadsCSSAndFeatures(t *testing.T) {
	org, err := Parse([]byte(`
title: Styled
style:
  backgroundColor: "#f8fafc"
  css: |
    div.person { border-top-color: teal; }
  features:
    avatars: false
    team: true
people:
  - name: Alice
`))
	require.NoError(t, err)

	features := org.Style.Features.Resolve()
	require.Equal(t, "#f8fafc", org.Style.BackgroundColor)
	require.Equal(t, "div.person { border-top-color: teal; }\n", org.Style.CSS)
	require.False(t, features.Avatars)
	require.True(t, features.JobTitles)
	require.True(t, features.Team)
	require.True(t, features.ExpandButtons)
}

func TestLoadDirUsesFileNameAsID(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Acme.yml"), []byte(sampleOrg), 0o644))

	loaded, err := LoadDir(dir)
	require.NoError(t, err)
	require.Equal(t, "Acme", loaded[0].ID)
}

func TestPersonAvatarURL(t *testing.T) {
	require.Equal(t, "/avatars/alice.jpg", (&Person{Name: "Alice Smith"}).AvatarURL())
	require.Equal(t, "/avatars/dana.png", (&Person{Name: "Dana", Avatar: "dana.png"}).AvatarURL())
	require.Equal(t, "/avatars/x.png", (&Person{Name: "Dana", Avatar: "../../x.png"}).AvatarURL())
}

func TestValidChartID(t *testing.T) {
	for _, id := range []string{"Acme", "acme-2026", "Second_Company"} {
		require.True(t, ValidChartID(id), id)
	}

	for _, id := range []string{"", "../etc", "a b", "a/b", "-acme", strings.Repeat("a", 65)} {
		require.False(t, ValidChartID(id), id)
	}
}

func TestWarningsListUnknownManagers(t *testing.T) {
	org, err := Parse([]byte(`
people:
  - name: Alice
    alias: ali
  - name: Bob
    reports: ali
  - name: Carol
    reports: Nobody
`))
	require.NoError(t, err)
	require.Equal(t, []string{`Carol reports to "Nobody", which matches no name or alias`}, org.Warnings())
}
