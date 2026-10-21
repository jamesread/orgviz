package httpserver

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func spaHandler(dir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := strings.TrimPrefix(r.URL.Path, "/")
		candidate := filepath.Join(dir, filepath.Clean(rel))

		if serveExistingFile(w, r, candidate) {
			return
		}

		if isMissingAsset(rel) {
			http.NotFound(w, r)
			return
		}

		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	})
}

func serveExistingFile(w http.ResponseWriter, r *http.Request, candidate string) bool {
	info, err := os.Stat(candidate)
	if err != nil || info.IsDir() {
		return false
	}

	http.ServeFile(w, r, candidate)
	return true
}

func isMissingAsset(rel string) bool {
	return rel != "" && filepath.Ext(rel) != ""
}
