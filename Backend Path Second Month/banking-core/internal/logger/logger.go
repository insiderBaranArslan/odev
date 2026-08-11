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

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}
