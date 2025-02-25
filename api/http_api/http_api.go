package http_api

import (
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"go_video_streamer/api/client_stream_creator"
	_ "go_video_streamer/docs"
	"log/slog"
	"net/http"
	"strings"
)

func mount(r *mux.Router, path string, handler http.Handler) {
	r.PathPrefix(path).Handler(
		http.StripPrefix(
			strings.TrimSuffix(path, "/"),
			handler,
		),
	)
}

func enableCORS(router *mux.Router) {
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
}

func BuildHttpAPI(enableStreamCreator bool, enableSwagger bool) *mux.Router {
	r := mux.NewRouter()
	enableCORS(r)

	prefix := "/api/v1"

	if enableStreamCreator {
		slog.Info("Mounting stream creator API", "prefix", prefix)
		mount(r, prefix, client_stream_creator.Router(prefix))
	}

	if enableSwagger {
		slog.Info("Mounting Swagger UI", "path", "/swagger/")
		r.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
			httpSwagger.URL("/docs/swagger.json"),
		))

		r.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs"))))
	}

	return r
}
