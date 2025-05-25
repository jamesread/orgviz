package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/jamesread/orgviz/internal/httpserver"
	"github.com/jamesread/orgviz/internal/config"
)

func main() {
	log.Infof("Starting orgviz")

	cfg := config.Get()

	httpserver.Start(cfg)
}

