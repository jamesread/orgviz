package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/jamesread/orgviz/internal/config"
	"github.com/jamesread/orgviz/internal/httpserver"
	log "github.com/sirupsen/logrus"
)

func main() {
	configDir := flag.String("configdir", "", "Directory containing config.yaml, orgs/, avatars/ and themes/")
	flag.Parse()

	log.Info("Starting orgviz")

	cfg, err := config.LoadFrom(*configDir)
	if err != nil {
		log.WithError(err).Fatal("Failed to load config")
	}

	logConfigFile(cfg)
	applyLogLevel(cfg.LogLevel)
	httpserver.Start(cfg, httpserver.Discovery{LLMsTxt: llmsTxt, OpenAPI: openAPISpec})
}

func logConfigFile(cfg *config.Config) {
	for _, searched := range cfg.SearchedPaths {
		log.WithField("path", absPath(searched)).Info("Searched for config file")
	}

	if cfg.ConfigFile == "" {
		log.WithField("configDir", cfg.ConfigDir).Info("No config file found, using defaults")
		return
	}

	log.WithField("file", absPath(cfg.ConfigFile)).Info("Loaded config file")
}

// absPath makes a path absolute for logging, keeping it as given on error.
func absPath(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}

	return path
}

func applyLogLevel(level string) {
	parsed, err := log.ParseLevel(level)
	if err != nil {
		log.WithField("logLevel", level).Warn("Unknown log level, using info")
		return
	}

	log.SetLevel(parsed)
}

func init() {
	log.SetOutput(os.Stderr)
}
