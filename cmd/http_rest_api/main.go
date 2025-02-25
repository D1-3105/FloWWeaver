package main

import (
	"fmt"
	"go_video_streamer/api/http_api"
	"go_video_streamer/cmd/utils"
	"log/slog"
	"net/http"
	"os"
)

// @title Video Streamer API
// @version 1.0
// @description API for video streaming service
// @BasePath /api/v1

func main() {
	utils.SetupSlog()
	port, exists := os.LookupEnv("FLOWWEAVER_REST_PORT")
	if !exists {
		port = "8081"
	}
	err := http.ListenAndServe(
		fmt.Sprintf(":%s", port),
		http_api.BuildHttpAPI(true, true),
	)
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err.Error()))
		os.Exit(1)
	}
}
