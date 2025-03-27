package internal

import (
	"context"
	"fmt"
	"log/slog"
)

// customHandler is a wrapper around slog.Handler that prepends a message
type customHandler struct {
	handler slog.Handler
	prepend string
}

func (h *customHandler) Handle(ctx context.Context, record slog.Record) error {
	record.Message = fmt.Sprintf("[%s] %s", h.prepend, record.Message)
	return h.handler.Handle(ctx, record)
}

// Enabled delegates to the underlying handler
func (h *customHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// WithAttrs delegates to the underlying handler
func (h *customHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &customHandler{
		handler: h.handler.WithAttrs(attrs),
		prepend: h.prepend,
	}
}

// WithGroup delegates to the underlying handler
func (h *customHandler) WithGroup(name string) slog.Handler {
	return &customHandler{
		handler: h.handler.WithGroup(name),
		prepend: h.prepend,
	}
}

func NewLogger(prepend string) *slog.Logger {
	h := &customHandler{
		handler: slog.Default().Handler(),
		prepend: prepend,
	}

	return slog.New(h)
}
