package log

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

const (
	LevelTrace slog.Level = -6
	LevelDebug            = slog.LevelDebug
	LevelInfo             = slog.LevelInfo
	LevelWarn             = slog.LevelWarn
	LevelError            = slog.LevelError
)

var ErrParsingLevel = errors.New("failed to parse level")

func ParseLevel(level string) (slog.Level, error) {
	level = strings.ToLower(level)
	switch level {
	case "":
		return LevelInfo, nil
	case "trace":
		return LevelTrace, nil
	case "debug":
		return LevelDebug, nil
	case "info":
		return LevelInfo, nil
	case "warn":
		return LevelWarn, nil
	case "error":
		return LevelError, nil
	}
	if i, err := strconv.Atoi(level); err == nil {
		return slog.Level(i), nil
	}
	return LevelInfo, ErrParsingLevel
}

func New(level slog.Level) *slog.Logger {
	levelVar := &slog.LevelVar{}
	levelVar.Set(level)

	opts := &slog.HandlerOptions{
		AddSource:   level <= LevelTrace,
		Level:       levelVar,
		ReplaceAttr: nil,
	}
	handler := slog.NewTextHandler(os.Stderr, opts)
	return NewWithHandler(handler)
}

func NewWithHandler(handler slog.Handler) *slog.Logger {
	return slog.New(&CtxHandler{handler})
}

type logCtxKey string

const (
	enableLevelKey logCtxKey = "enableLevel"
	attrsKey       logCtxKey = "attrs"
	levelAttrsKey  logCtxKey = "levelAttrs"
)

func WithLevel(ctx context.Context, level slog.Level) context.Context {
	return context.WithValue(ctx, enableLevelKey, level)
}

type CtxHandler struct {
	slog.Handler
}

func (c *CtxHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if ctxLevel, ok := ctx.Value(enableLevelKey).(slog.Level); ok {
		return level >= ctxLevel
	}
	return c.Handler.Enabled(ctx, level)
}

func (c *CtxHandler) Handle(ctx context.Context, record slog.Record) error {
	var attrs []slog.Attr
	ctxLevelAttrs := ctxLevelAttrs(ctx)
	for _, levelAttrs := range ctxLevelAttrs[:min(len(ctxLevelAttrs), levelToIdx(record.Level)+1)] {
		attrs = append(attrs, levelAttrs...)
	}
	attrs = append(attrs, ctxAttrs(ctx)...)
	return c.Handler.WithAttrs(attrs).Handle(ctx, record)
}

func ctxAttrs(ctx context.Context) []slog.Attr {
	attrs, _ := ctx.Value(attrsKey).([]slog.Attr)
	return attrs
}

func WithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	ctxAttrs := ctxAttrs(ctx)
	return context.WithValue(ctx, attrsKey, append(ctxAttrs, attrs...))
}

func WithArgs(ctx context.Context, args ...any) context.Context {
	if len(args) == 0 {
		return ctx
	}
	return WithAttrs(ctx, argsToAttrSlice(args)...)
}

const (
	// based on opentelemetry levels, which is part of the basis of slog levels
	// https://pkg.go.dev/log/slog#Level
	minLevel    slog.Level = -9
	maxLevelIdx int        = 24
)

func levelToIdx(level slog.Level) int {
	// Invert the levels, so we can subslice from the beginning to get levels >= the requested level
	return maxLevelIdx - int(level-minLevel)
}

func ctxLevelAttrs(ctx context.Context) [][]slog.Attr {
	attrs, _ := ctx.Value(levelAttrsKey).([][]slog.Attr)
	return attrs
}

func WithLevelAttrs(ctx context.Context, level slog.Level, attrs ...slog.Attr) context.Context {
	ctxAttrs := ctxLevelAttrs(ctx)
	i := levelToIdx(level)
	if i >= len(ctxAttrs) {
		ctxAttrs = append(ctxAttrs, make([][]slog.Attr, i+1-len(ctxAttrs))...)
	}
	ctxAttrs[i] = append(ctxAttrs[i], attrs...)
	return context.WithValue(ctx, levelAttrsKey, ctxAttrs)
}

func WithLevelArgs(ctx context.Context, level slog.Level, args ...any) context.Context {
	if len(args) == 0 {
		return ctx
	}
	return WithLevelAttrs(ctx, level, argsToAttrSlice(args)...)
}
