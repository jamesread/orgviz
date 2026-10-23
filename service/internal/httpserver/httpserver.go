package httpserver

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	connectcors "connectrpc.com/cors"
	"github.com/jamesread/orgviz/gen/orgviz/clientapi/v1/clientapiconnect"
	"github.com/jamesread/orgviz/internal/charts"
	"github.com/jamesread/orgviz/internal/clientapi"
	"github.com/jamesread/orgviz/internal/config"
	"github.com/jamesread/orgviz/internal/mcpserver"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

func withCors(handler http.Handler) http.Handler {
	mw := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: connectcors.ExposedHeaders(),
	})

	return mw.Handler(handler)
}

func listenAddr(cfg *config.Config) string {
	if addr := strings.TrimSpace(os.Getenv("ORGVIZ_LISTEN")); addr != "" {
		return addr
	}

	port := "8080"
	if cfg != nil && cfg.Port != "" {
		port = cfg.Port
	}

	return ":" + port
}

func findFirstExistingDirectory(name string, directoriesToSearch []string) string {
	for _, dir := range directoriesToSearch {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			absdir, _ := filepath.Abs(dir)
			log.WithFields(log.Fields{"dir": dir, "absdir": absdir}).Infof("Found the %v directory", name)
			return absdir
		}
	}

	log.Warnf("Did not find the %v directory, you will probably get 404 errors.", name)
	return "./null/"
}

func findWebuiDir() string {
	if envDir := strings.TrimSpace(os.Getenv("ORGVIZ_FRONTEND_DIR")); envDir != "" {
		return envDir
	}

	return findFirstExistingDirectory("webui", []string{
		"../frontend/dist/",
		"../frontend-dist/",
		"../frontend/",
		"/frontend/",
		"/usr/share/orgviz/frontend/",
		"/var/www/orgviz/",
		"/etc/orgviz/frontend/",
	})
}

func findAvatarDir(cfg *config.Config) string {
	candidates := []string{
		"../var/config-skel/avatars/",
		"/config/avatars/",
		"/usr/share/orgviz/avatars/",
		"/var/www/orgviz/avatars/",
		"/etc/orgviz/avatars/",
	}

	if cfg != nil && cfg.AvatarsDir() != "" {
		candidates = append([]string{cfg.AvatarsDir()}, candidates...)
	}

	return findFirstExistingDirectory("avatars", candidates)
}

func findOrgsDir(cfg *config.Config) string {
	candidates := []string{
		"../var/config-skel/orgs/",
		"/config/orgs/",
		"../examples/",
	}

	if cfg != nil && cfg.OrgsDir() != "" {
		candidates = append([]string{cfg.OrgsDir()}, candidates...)
	}

	return findFirstExistingDirectory("orgs", candidates)
}

func findThemesDir(cfg *config.Config) string {
	if cfg != nil && cfg.ConfigDir != "" {
		return cfg.ThemesDir()
	}

	return findFirstExistingDirectory("themes", []string{
		"../var/config-skel/themes/",
		"/config/themes/",
	})
}

// Discovery holds the documents that describe the API to AI agents and tools.
type Discovery struct {
	LLMsTxt []byte
	OpenAPI []byte
}

func serveBytes(contentType string, body []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		_, _ = w.Write(body)
	}
}

func mcpHandler(cfg *config.Config, store *charts.Store) http.Handler {
	if !cfg.MCPEnabled {
		log.Info("MCP server is disabled; set mcp.enabled: true in config.yaml or ORGVIZ_MCP_ENABLED=true to serve it at /mcp")
		return http.NotFoundHandler()
	}

	log.WithField("path", "/mcp").Warn("MCP server is enabled with no authentication: anyone who can reach orgviz can create and change charts, themes and avatars")

	return mcpserver.NewHandler(store)
}

func newMux(cfg *config.Config, discovery Discovery) *http.ServeMux {
	avatarDir := findAvatarDir(cfg)
	store := charts.New(findOrgsDir(cfg), avatarDir, findThemesDir(cfg))

	apiPath, apiHandler := clientapiconnect.NewClientToFrontendHandler(clientapi.NewServer(store))

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/api"+apiPath, func(w http.ResponseWriter, r *http.Request) {
		log.Infof("API request: %s", r.URL.Path)
		http.StripPrefix("/api", withCors(apiHandler)).ServeHTTP(w, r)
	})
	mux.HandleFunc("GET /llms.txt", serveBytes("text/plain; charset=utf-8", discovery.LLMsTxt))
	mux.HandleFunc("GET /openapi", serveBytes("application/json", discovery.OpenAPI))
	mux.Handle("/mcp", mcpHandler(cfg, store))
	mux.Handle("/avatars/", http.StripPrefix("/avatars/", http.FileServer(http.Dir(avatarDir))))
	mux.Handle("/", spaHandler(findWebuiDir()))

	return mux
}

// Start serves the API, avatars, metrics, discovery documents, MCP and web UI.
func Start(cfg *config.Config, discovery Discovery) {
	addr := listenAddr(cfg)
	log.WithField("addr", addr).Info("Starting HTTP server")

	srv := &http.Server{
		Addr:    addr,
		Handler: newMux(cfg, discovery),
	}

	log.Fatal(srv.ListenAndServe())
}
