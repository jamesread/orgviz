package main

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestApplyLogLevel(t *testing.T) {
	applyLogLevel("debug")
	require.Equal(t, log.DebugLevel, log.GetLevel())

	applyLogLevel("not-a-level")
	require.Equal(t, log.DebugLevel, log.GetLevel())

	applyLogLevel("info")
	require.Equal(t, log.InfoLevel, log.GetLevel())
}
