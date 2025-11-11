package log

import (
	"log/slog"
	"os"

	"github.com/phsym/console-slog"
)

func NewConsoleHandler(level slog.Level) *slog.Logger {
	return NewWithHandler(
		console.NewHandler(os.Stderr, &console.HandlerOptions{Level: level, AddSource: level <= LevelTrace, TimeFormat: ""}),
	)
}
