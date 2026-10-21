package orgs

import (
	"os"
	"path/filepath"
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
