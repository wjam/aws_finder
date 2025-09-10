package log

import (
	"context"
	"log/slog"
	"strings"
)

type attrsKey struct{}
type loggerKey struct{}

func WithAttrs(ctx context.Context, attr ...slog.Attr) context.Context {
	var existing []slog.Attr
	if v := ctx.Value(attrsKey{}); v != nil {
		existing = v.([]slog.Attr)
	}
	return context.WithValue(ctx, attrsKey{}, append(existing, attr...))
}

func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func Logger(ctx context.Context) *slog.Logger {
	return ctx.Value(loggerKey{}).(*slog.Logger)
}

var _ slog.Handler = WithAttrsFromContextHandler{}

type WithAttrsFromContextHandler struct {
	Parent slog.Handler
}

func (w WithAttrsFromContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return w.Parent.Enabled(ctx, level)
}

func (w WithAttrsFromContextHandler) Handle(ctx context.Context, record slog.Record) error {
	if v := ctx.Value(attrsKey{}); v != nil {
		record.AddAttrs(v.([]slog.Attr)...)
	}

	return w.Parent.Handle(ctx, record)
}

func (w WithAttrsFromContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return w.Parent.WithAttrs(attrs)
}

func (w WithAttrsFromContextHandler) WithGroup(name string) slog.Handler {
	return w.Parent.WithGroup(name)
}

func FilterAttributesFromLog(ignored []string) func(groups []string, a slog.Attr) slog.Attr {
	lookup := make(map[string]struct{}, len(ignored))
	for _, s := range ignored {
		lookup[s] = struct{}{}
	}
	return func(groups []string, a slog.Attr) slog.Attr {
		parts := append(groups, a.Key) //nolint:gocritic // two slices are semantically different
		for i := 1; i <= len(parts); i++ {
			key := strings.Join(parts[:i], ".")
			if _, ok := lookup[key]; ok {
				return slog.Attr{}
			}
		}
		return a
	}
}
