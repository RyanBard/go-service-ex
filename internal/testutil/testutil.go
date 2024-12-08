package testutil

import (
	"log/slog"
	"os"
)

func GetLogger() *slog.Logger {
	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelError)
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}
