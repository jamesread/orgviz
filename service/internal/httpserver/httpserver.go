package httpserver

import (
	"github.com/jamesread/orgviz/internal/config"
	"github.com/jamesread/orgviz/internal/clientapi"
	"github.com/jamesread/orgviz/gen/orgviz/clientapi/v1/clientapiconnect"

	connectcors "connectrpc.com/cors"

	"fmt"

	"github.com/rs/cors"

	"path/filepath"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

func withCors(h http.Handler) http.Handler {
	mw := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   connectcors.AllowedMethods(),
		AllowedHeaders:   connectcors.AllowedHeaders(),
		ExposedHeaders:   connectcors.ExposedHeaders(),
	})

	return mw.Handler(h)
}

func getNewApiHandler(cfg *config.Config) (string, http.Handler) {
	apiServer := clientapi.NewServer()

	path, handler := clientapiconnect.NewClientToFrontendHandler(apiServer)

	return path, withCors(handler)
}

func findWebuiDir() string {
	return findFirstExistingDirectory("webui", []string{
		"../frontend/dist/",
		"../frontend/",
		"/frontend/",
		"/usr/share/orgviz/frontend/",
		"/var/www/orgviz/",
		"/etc/orgviz/frontend/",
	})
}

func findFirstExistingDirectory(name string, directoriesToSearch []string) string {
	for _, dir := range directoriesToSearch {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			absdir, _ := filepath.Abs(dir)

			log.WithFields(log.Fields{
				"dir": dir,
				"absdir": absdir,
			}).Infof("Found the %v directory", name)

			return absdir
		}
	}

	log.Warnf("Did not find the %v directory, you will probably get 404 errors.", name)

	return "./null/" // Should not exist
}

func findAvatarDir() string {
	return findFirstExistingDirectory("avatars", []string{
		"../var/config-skel/avatars/",
		"/config/avatars/",
		"/usr/share/orgviz/avatars/",
		"/var/www/orgviz/avatars/",
		"/etc/orgviz/avatars/",
	})
}

func getNewWebUIHandler(dir string) http.Handler {
	return http.FileServer(http.Dir(dir))
}

func Start(cfg *config.Config) {
	const port = 8080

	log.WithFields(log.Fields{
		"port": port,
	}).Info("Starting HTTP server")

	apipath, apihandler := getNewApiHandler(cfg)

	log.Infof("API path: %s", apipath)

	mux := http.NewServeMux()

	mux.HandleFunc("/api"+apipath, func(w http.ResponseWriter, r *http.Request) {
		log.Infof("API request: %s", r.URL.Path)

		http.StripPrefix("/api", apihandler).ServeHTTP(w, r)
	})

	mux.Handle("/avatars/", http.StripPrefix("/avatars/", http.FileServer(http.Dir(findAvatarDir()))))
	mux.Handle("/", getNewWebUIHandler(findWebuiDir()))

	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%v", port),
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}
