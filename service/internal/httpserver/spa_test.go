package httpserver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpaHandlerServesIndexForClientRoutes(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "index.html"), []byte("spa"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "logo.png"), []byte("img"), 0o644))

	handler := spaHandler(dir)

	index := httptest.NewRecorder()
	handler.ServeHTTP(index, httptest.NewRequest(http.MethodGet, "/chart/0", nil))
	require.Equal(t, http.StatusOK, index.Code)
	require.Equal(t, "spa", index.Body.String())

	asset := httptest.NewRecorder()
	handler.ServeHTTP(asset, httptest.NewRequest(http.MethodGet, "/logo.png", nil))
	require.Equal(t, http.StatusOK, asset.Code)
	require.Equal(t, "img", asset.Body.String())

	missing := httptest.NewRecorder()
	handler.ServeHTTP(missing, httptest.NewRequest(http.MethodGet, "/missing.js", nil))
	require.Equal(t, http.StatusNotFound, missing.Code)
}
