package main

import (
	"context"
	"errors"
	"fmt"
	"go_video_streamer/api/grpc_consumer"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	serverPort := os.Getenv("GRPC_PORT")
	go func(ctx context.Context) {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", serverPort))
		if err != nil {
			panic(err)
		}
		grpcServer := grpc.NewServer()
		grpc_consumer.RegisterVideoRCVServer(grpcServer, grpc_consumer.NewGRPCVideoRCVServer())
		err = grpcServer.Serve(lis)
		if err != nil {
			slog.Error(fmt.Sprintf("failed to serve: %s", err))
		}
	}(ctx)

	if os.Getenv("SERVER_OFF") != "true" {
		go func(ctx context.Context) {
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
			}
		}(ctx)
	}
	select {
	case <-ctx.Done():
		slog.Info("Server gracefully closed.")
		break
	case <-sigs:
		cancel()
		slog.Info("Server closed.")
		break
	}
}
