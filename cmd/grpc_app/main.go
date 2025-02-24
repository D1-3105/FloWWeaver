package main

import (
	"context"
	"fmt"
	"go_video_streamer/api/file_server"
	"go_video_streamer/api/grpc_consumer"
	"google.golang.org/grpc"
	"log/slog"
	"net"
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
		go file_server.LaunchStreamRepoFileServer(ctx)
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
