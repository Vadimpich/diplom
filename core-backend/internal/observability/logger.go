package observability

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

func NewLogger(level string, writer io.Writer) *slog.Logger {
	if writer == nil {
		writer = os.Stdout
	}
	var slogLevel slog.Level
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "DEBUG":
		slogLevel = slog.LevelDebug
	case "WARN":
		slogLevel = slog.LevelWarn
	case "ERROR":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: slogLevel}))
}
