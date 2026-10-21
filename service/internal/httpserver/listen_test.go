package httpserver

import (
	"testing"

	"github.com/jamesread/orgviz/internal/config"
	"github.com/stretchr/testify/require"
)

func TestListenAddrDefault(t *testing.T) {
	t.Setenv("ORGVIZ_LISTEN", "")
	require.Equal(t, ":8080", listenAddr(&config.Config{Port: "8080"}))
}

func TestListenAddrUsesPORTFromConfig(t *testing.T) {
	t.Setenv("ORGVIZ_LISTEN", "")
	require.Equal(t, ":18480", listenAddr(&config.Config{Port: "18480"}))
}

func TestListenAddrORGVIZListenWins(t *testing.T) {
	t.Setenv("ORGVIZ_LISTEN", "127.0.0.1:18480")
	require.Equal(t, "127.0.0.1:18480", listenAddr(&config.Config{Port: "8080"}))
}
