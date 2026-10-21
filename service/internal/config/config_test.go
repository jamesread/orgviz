package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadFromDefaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("ORGVIZ_CONFIGDIR", "")

	cfg, err := LoadFrom("")
	require.NoError(t, err)
	require.Equal(t, "8080", cfg.Port)
	require.Equal(t, "info", cfg.LogLevel)
	require.True(t, cfg.ShowFooter)
}

func TestLoadFromPORTEnv(t *testing.T) {
	t.Setenv("PORT", "18480")
	t.Setenv("ORGVIZ_CONFIGDIR", "")

	cfg, err := LoadFrom("")
	require.NoError(t, err)
	require.Equal(t, "18480", cfg.Port)
}

func TestLoadFromConfigFile(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("ORGVIZ_CONFIGDIR", "")

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(""+
		"configVersion: 1\n"+
		"logLevel: debug\n"+
		"showFooter: false\n"+
		"listenPort: \"9090\"\n"), 0o644))

	cfg, err := LoadFrom(dir)
	require.NoError(t, err)
	require.Equal(t, dir, cfg.ConfigDir)
	require.Equal(t, "debug", cfg.LogLevel)
	require.False(t, cfg.ShowFooter)
	require.Equal(t, "9090", cfg.Port)
}

func TestLoadFromPORTOverridesFile(t *testing.T) {
	t.Setenv("PORT", "18480")

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("listenPort: \"9090\"\n"), 0o644))

	cfg, err := LoadFrom(dir)
	require.NoError(t, err)
	require.Equal(t, "18480", cfg.Port)
}

func TestOrgsDirPrefersConfigDir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "orgs"), 0o755))

	cfg := &Config{ConfigDir: dir}
	require.Equal(t, filepath.Join(dir, "orgs"), cfg.OrgsDir())
}

func TestAvatarsDirPrefersConfigDir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "avatars"), 0o755))

	cfg := &Config{ConfigDir: dir}
	require.Equal(t, filepath.Join(dir, "avatars"), cfg.AvatarsDir())
}
