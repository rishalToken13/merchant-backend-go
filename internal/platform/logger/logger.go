package logger

import (
	"log/slog"
	"os"
)

type Options struct {
	Env string // dev/prod
}

func New(opts Options) *slog.Logger {
	level := slog.LevelInfo
	if opts.Env == "dev" {
		level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	return slog.New(handler)
}
