package logger

import (
	"log/slog"
	"os"
	"strings"
)

func New(env string) *slog.Logger {
	level := slog.LevelInfo
	if strings.EqualFold(env, "development") {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)
	return logger
}
