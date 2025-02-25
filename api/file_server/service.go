package file_server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

func BuildHLSPlaylistURL(streamName string) string {
	return fmt.Sprintf("/hls/%s/index.m3u8", streamName)
}

func LaunchStreamRepoFileServer(_ context.Context) {
	http.HandleFunc("/hls/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			return
		}

		http.StripPrefix("/hls/", http.FileServer(http.Dir("./stream_repo/"))).ServeHTTP(w, r)
	})

	err := http.ListenAndServe(":8080", nil)
	if errors.Is(err, http.ErrServerClosed) {
		slog.Info("server closed")
	} else if err != nil {
		slog.Error("error starting server: %s", err)
		os.Exit(1)
	}
}
