package logger

import (
	"context"
	"log/slog"
	"os"
)

var defaultLogger *slog.Logger
var logLevel = new(slog.LevelVar)

const (
	LevelDebug = int(slog.LevelDebug)
	LevelInfo  = int(slog.LevelInfo)
	LevelWarn  = int(slog.LevelWarn)
	LevelError = int(slog.LevelError)
)

func Init() {
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	// Use JSON handler for structured logging
	handler := slog.NewJSONHandler(os.Stdout, opts)
	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

func SetLogLevel(level int) {
	logLevel.Set(slog.Level(level))
}


func Debug(tag string, msg string, args ...any) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.Debug(msg, append([]any{slog.String("tag", tag)}, args...)...)
}

func Info(tag string, msg string, args ...any) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.Info(msg, append([]any{slog.String("tag", tag)}, args...)...)
}

func Warn(tag string, msg string, args ...any) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.Warn(msg, append([]any{slog.String("tag", tag)}, args...)...)
}

func Error(tag string, msg string, args ...any) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.Error(msg, append([]any{slog.String("tag", tag)}, args...)...)
}

func Panic(tag string, msg string, args ...any) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.Error(msg, append([]any{slog.String("tag", tag)}, args...)...)
	os.Exit(1)
}

// Context aware logging
func DebugContext(ctx context.Context, msg string, args ...any) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.DebugContext(ctx, msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.InfoContext(ctx, msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.ErrorContext(ctx, msg, args...)
}
