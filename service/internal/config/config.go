package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const envConfigDir = "ORGVIZ_CONFIGDIR"

// Config holds application configuration loaded via koanf.
type Config struct {
	Port       string
	LogLevel   string
	ShowFooter bool
	ConfigDir  string
}

// LoadFrom builds configuration from a config directory, config.yaml, and env.
func LoadFrom(configDir string) (*Config, error) {
	dir := resolveConfigDir(configDir)
	k := koanf.New(".")

	if err := loadConfigFile(k, dir); err != nil {
		return nil, err
	}

	cfg := &Config{
		Port:       firstNonEmpty(os.Getenv("PORT"), k.String("listenPort"), "8080"),
		LogLevel:   firstNonEmpty(k.String("logLevel"), "info"),
		ShowFooter: true,
		ConfigDir:  dir,
	}

	if k.Exists("showFooter") {
		cfg.ShowFooter = k.Bool("showFooter")
	}

	return cfg, nil
}

func resolveConfigDir(flagDir string) string {
	if flagDir != "" {
		return flagDir
	}

	if envDir := os.Getenv(envConfigDir); envDir != "" {
		return envDir
	}

	if dirExists("/config") {
		return "/config"
	}

	return ""
}

func loadConfigFile(k *koanf.Koanf, dir string) error {
	if dir == "" {
		return nil
	}

	configPath := filepath.Join(dir, "config.yaml")
	if !fileExists(configPath) {
		return nil
	}

	if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		return fmt.Errorf("loading config file: %w", err)
	}

	return nil
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
