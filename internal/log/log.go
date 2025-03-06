package log

import (
	"context"
	"log/slog"
	"slices"
	"time"
)

type attrsKey struct{}
type loggerKey struct{}

func WithAttrs(ctx context.Context, attr ...slog.Attr) context.Context {
	var existing []slog.Attr
	if v := ctx.Value(attrsKey{}); v != nil {
		existing = v.([]slog.Attr)
	}
	return context.WithValue(ctx, attrsKey{}, append(slices.Clone(existing), attr...))
}

func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func Logger(ctx context.Context) *slog.Logger {
	return ctx.Value(loggerKey{}).(*slog.Logger)
}

var _ slog.Handler = WithAttrsFromContextHandler{}

type WithAttrsFromContextHandler struct {
	Parent            slog.Handler
	IgnoredAttributes []string
}

func (w WithAttrsFromContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return w.Parent.Enabled(ctx, level)
}

func (w WithAttrsFromContextHandler) Handle(ctx context.Context, record slog.Record) error {
	if v := ctx.Value(attrsKey{}); v != nil {
		record.AddAttrs(v.([]slog.Attr)...)
	}

	newRecord := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)

	if slices.Contains(w.IgnoredAttributes, "time") {
		newRecord.Time = time.Time{}
	}

	record.Attrs(func(a slog.Attr) bool {
		if slices.Contains(w.IgnoredAttributes, a.Key) {
			return true
		}

		newRecord.AddAttrs(a)
		return true
	})

	return w.Parent.Handle(ctx, newRecord)
}

func (w WithAttrsFromContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return w.Parent.WithAttrs(attrs)
}

func (w WithAttrsFromContextHandler) WithGroup(name string) slog.Handler {
	return w.Parent.WithGroup(name)
}
