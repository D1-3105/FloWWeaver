package utils

import (
	"log/slog"
	"os"
	"strconv"
)

func SetupSlog() {
	level, exists := os.LookupEnv("SLOG_LEVEL")
	if !exists {
		level = "0"
	}
	levelInt, err := strconv.Atoi(level)
	if err != nil {
		panic(err)
	}
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.Level(levelInt),
	})
	slog.SetDefault(slog.New(h))
}
