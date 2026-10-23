package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const (
	envConfigDir  = "ORGVIZ_CONFIGDIR"
	envMCPEnabled = "ORGVIZ_MCP_ENABLED"
)

// Config holds application configuration loaded via koanf.
type Config struct {
	Port       string
	LogLevel   string
	ShowFooter bool
	ConfigDir  string

	// ConfigFile is the config.yaml that was loaded, or empty when none was.
	ConfigFile string

	// SearchedPaths lists the config.yaml paths that were looked for, in order.
	SearchedPaths []string

	// MCPEnabled serves the MCP endpoint at /mcp. It has no authentication
	// and can write org files and avatars, so it is off by default.
	MCPEnabled bool
}

// LoadFrom builds configuration from a config directory, config.yaml, and env.
func LoadFrom(configDir string) (*Config, error) {
	dir, searched := resolveConfigDir(configDir)
	k := koanf.New(".")

	configFile, err := loadConfigFile(k, dir)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Port:       firstNonEmpty(os.Getenv("PORT"), k.String("listenPort"), "8080"),
		LogLevel:   firstNonEmpty(k.String("logLevel"), "info"),
		ShowFooter: true,
		ConfigDir:  dir,
		ConfigFile: configFile,

		SearchedPaths: searched,
	}

	if k.Exists("showFooter") {
		cfg.ShowFooter = k.Bool("showFooter")
	}

	cfg.MCPEnabled = k.Bool("mcp.enabled") || envBool(envMCPEnabled)

	return cfg, nil
}

// resolveConfigDir picks the config directory: the -configdir flag, then
// ORGVIZ_CONFIGDIR, then the first of searchedConfigDirs that exists. It also
// returns the config.yaml paths that were searched, in order.
func resolveConfigDir(flagDir string) (string, []string) {
	if dir := firstNonEmpty(flagDir, os.Getenv(envConfigDir)); dir != "" {
		return dir, []string{filepath.Join(dir, "config.yaml")}
	}

	searched := []string{}
	for _, dir := range searchedConfigDirs() {
		searched = append(searched, filepath.Join(dir, "config.yaml"))
		if dirExists(dir) {
			return dir, searched
		}
	}

	return "", searched
}

// searchedConfigDirs is /config (for containers) and then orgviz in the user's
// config directory, which is $XDG_CONFIG_HOME or ~/.config on Linux.
func searchedConfigDirs() []string {
	dirs := []string{"/config"}

	if userDir, err := os.UserConfigDir(); err == nil {
		dirs = append(dirs, filepath.Join(userDir, "orgviz"))
	}

	return dirs
}

// loadConfigFile loads dir/config.yaml if it exists and returns its path, or
// an empty path when there is no file to load.
func loadConfigFile(k *koanf.Koanf, dir string) (string, error) {
	if dir == "" {
		return "", nil
	}

	configPath := filepath.Join(dir, "config.yaml")
	if !fileExists(configPath) {
		return "", nil
	}

	if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		return "", fmt.Errorf("loading config file: %w", err)
	}

	return configPath, nil
}

func (cfg *Config) OrgsDir() string {
	if cfg == nil || cfg.ConfigDir == "" {
		return ""
	}

	return filepath.Join(cfg.ConfigDir, "orgs")
}

func (cfg *Config) AvatarsDir() string {
	if cfg == nil || cfg.ConfigDir == "" {
		return ""
	}

	return filepath.Join(cfg.ConfigDir, "avatars")
}

func (cfg *Config) ThemesDir() string {
	if cfg == nil || cfg.ConfigDir == "" {
		return ""
	}

	return filepath.Join(cfg.ConfigDir, "themes")
}

func envBool(name string) bool {
	value, err := strconv.ParseBool(os.Getenv(name))
	return err == nil && value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
