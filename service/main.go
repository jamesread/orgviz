package main

import (
	"flag"
	"os"

	"github.com/jamesread/orgviz/internal/config"
	"github.com/jamesread/orgviz/internal/httpserver"
	log "github.com/sirupsen/logrus"
)

func main() {
	configDir := flag.String("configdir", "", "Directory containing config.yaml, orgs/ and avatars/")
	flag.Parse()

	log.Info("Starting orgviz")

	cfg, err := config.LoadFrom(*configDir)
	if err != nil {
		log.WithError(err).Fatal("Failed to load config")
	}

	applyLogLevel(cfg.LogLevel)
	httpserver.Start(cfg)
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
