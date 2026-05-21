package internal

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

var logger *slog.Logger

func initLogger(toFile bool) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{
					Key:   slog.TimeKey,
					Value: slog.StringValue(a.Value.Time().Format(time.RFC3339)),
				}
			}
			return a
		},
	}

	var writers []io.Writer
	writers = append(writers, os.Stdout)

	if toFile {
		logPath := filepath.Join("passman.log")
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			writers = append(writers, f)
		} else {
			fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
		}
	}

	multi := io.MultiWriter(writers...)
	handler := slog.NewTextHandler(multi, opts)
	logger = slog.New(handler)
	slog.SetDefault(logger)
}
