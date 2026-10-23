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
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := LoadFrom("")
	require.NoError(t, err)
	require.Equal(t, "8080", cfg.Port)
	require.Equal(t, "info", cfg.LogLevel)
	require.True(t, cfg.ShowFooter)
	require.Equal(t, "", cfg.ConfigFile)
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
	require.Equal(t, filepath.Join(dir, "config.yaml"), cfg.ConfigFile)
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

func TestThemesDirPrefersConfigDir(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{ConfigDir: dir}
	require.Equal(t, filepath.Join(dir, "themes"), cfg.ThemesDir())
}

func TestMCPDisabledByDefault(t *testing.T) {
	t.Setenv("ORGVIZ_CONFIGDIR", "")
	t.Setenv("ORGVIZ_MCP_ENABLED", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := LoadFrom("")
	require.NoError(t, err)
	require.False(t, cfg.MCPEnabled)
}

func TestMCPEnabledFromConfigFile(t *testing.T) {
	t.Setenv("ORGVIZ_CONFIGDIR", "")
	t.Setenv("ORGVIZ_MCP_ENABLED", "")

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("mcp:\n  enabled: true\n"), 0o644))

	cfg, err := LoadFrom(dir)
	require.NoError(t, err)
	require.True(t, cfg.MCPEnabled)
}

func TestMCPEnabledFromEnv(t *testing.T) {
	t.Setenv("ORGVIZ_CONFIGDIR", "")
	t.Setenv("ORGVIZ_MCP_ENABLED", "true")

	cfg, err := LoadFrom("")
	require.NoError(t, err)
	require.True(t, cfg.MCPEnabled)
}

func TestConfigFileEmptyWhenDirHasNoConfig(t *testing.T) {
	t.Setenv("ORGVIZ_CONFIGDIR", "")

	dir := t.TempDir()
	cfg, err := LoadFrom(dir)
	require.NoError(t, err)
	require.Equal(t, dir, cfg.ConfigDir)
	require.Equal(t, "", cfg.ConfigFile)
}

func TestSearchedPathsFromFlag(t *testing.T) {
	t.Setenv("ORGVIZ_CONFIGDIR", "/from/env")

	dir := t.TempDir()
	cfg, err := LoadFrom(dir)
	require.NoError(t, err)
	require.Equal(t, []string{filepath.Join(dir, "config.yaml")}, cfg.SearchedPaths)
}

func TestSearchedPathsFromEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ORGVIZ_CONFIGDIR", dir)

	cfg, err := LoadFrom("")
	require.NoError(t, err)
	require.Equal(t, []string{filepath.Join(dir, "config.yaml")}, cfg.SearchedPaths)
}

func TestSearchedPathsDefault(t *testing.T) {
	if dirExists("/config") {
		t.Skip("/config exists on this machine and ends the search")
	}

	t.Setenv("ORGVIZ_CONFIGDIR", "")
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", home)

	cfg, err := LoadFrom("")
	require.NoError(t, err)
	require.Equal(t, []string{"/config/config.yaml", filepath.Join(home, "orgviz", "config.yaml")}, cfg.SearchedPaths)
}

func TestConfigDirFallsBackToUserConfigDir(t *testing.T) {
	if dirExists("/config") {
		t.Skip("/config exists on this machine and is searched first")
	}

	t.Setenv("ORGVIZ_CONFIGDIR", "")
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", home)

	userDir := filepath.Join(home, "orgviz")
	require.NoError(t, os.MkdirAll(userDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(userDir, "config.yaml"), []byte("logLevel: debug\n"), 0o644))

	cfg, err := LoadFrom("")
	require.NoError(t, err)
	require.Equal(t, userDir, cfg.ConfigDir)
	require.Equal(t, filepath.Join(userDir, "config.yaml"), cfg.ConfigFile)
	require.Equal(t, "debug", cfg.LogLevel)
	require.Equal(t, []string{"/config/config.yaml", filepath.Join(userDir, "config.yaml")}, cfg.SearchedPaths)
}

func TestSearchedConfigDirsOrder(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/home/someone/.config")

	require.Equal(t, []string{"/config", "/home/someone/.config/orgviz"}, searchedConfigDirs())
}

func TestFlagAndEnvBeatSearchedDirs(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", home)
	require.NoError(t, os.MkdirAll(filepath.Join(home, "orgviz"), 0o755))

	t.Setenv("ORGVIZ_CONFIGDIR", "/from/env")

	dir, _ := resolveConfigDir("")
	require.Equal(t, "/from/env", dir)

	dir, _ = resolveConfigDir("/from/flag")
	require.Equal(t, "/from/flag", dir)
}
