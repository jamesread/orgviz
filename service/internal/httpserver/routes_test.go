package httpserver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesread/orgviz/internal/config"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

var testDiscovery = Discovery{
	LLMsTxt: []byte("# orgviz\n"),
	OpenAPI: []byte(`{"openapi":"3.1.0"}`),
}

func testConfig(t *testing.T, mcpEnabled bool) *config.Config {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "orgs"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "avatars"), 0o755))

	return &config.Config{ConfigDir: dir, MCPEnabled: mcpEnabled}
}

func get(t *testing.T, handler http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestDiscoveryRoutes(t *testing.T) {
	mux := newMux(testConfig(t, false), testDiscovery)

	llms := get(t, mux, "/llms.txt")
	require.Equal(t, http.StatusOK, llms.Code)
	require.Equal(t, "text/plain; charset=utf-8", llms.Header().Get("Content-Type"))
	require.Equal(t, "# orgviz\n", llms.Body.String())

	openapi := get(t, mux, "/openapi")
	require.Equal(t, http.StatusOK, openapi.Code)
	require.Equal(t, "application/json", openapi.Header().Get("Content-Type"))
	require.Contains(t, openapi.Body.String(), "3.1.0")
}

func TestMCPIsNotFoundWhenDisabled(t *testing.T) {
	mux := newMux(testConfig(t, false), testDiscovery)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader("{}")))
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestMCPAnswersWhenEnabled(t *testing.T) {
	mux := newMux(testConfig(t, true), testDiscovery)

	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"name":"orgviz"`)
}

func TestMCPStateIsLogged(t *testing.T) {
	hook := logtest.NewGlobal()
	defer hook.Reset()

	newMux(testConfig(t, false), testDiscovery)
	require.Contains(t, mcpLogMessages(hook), "MCP server is disabled")

	hook.Reset()
	newMux(testConfig(t, true), testDiscovery)
	require.Contains(t, mcpLogMessages(hook), "MCP server is enabled")
}

func mcpLogMessages(hook *logtest.Hook) string {
	messages := []string{}
	for _, entry := range hook.AllEntries() {
		if strings.Contains(entry.Message, "MCP") {
			messages = append(messages, entry.Message)
		}
	}

	return strings.Join(messages, "\n")
}
