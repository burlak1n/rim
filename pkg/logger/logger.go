package logger

import (
	"log/slog"
	"os"
)

// NewLogger создает и настраивает новый экземпляр slog.Logger.
// Уровень логирования определяется переменной окружения LOG_LEVEL (по умолчанию INFO).
func NewLogger() *slog.Logger {
	var logLevel slog.Level
	switch os.Getenv("LOG_LEVEL") {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return logger
}
