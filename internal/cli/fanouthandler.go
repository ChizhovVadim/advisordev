package cli

import (
	"context"
	"log/slog"
)

type FanoutHandler struct {
	handlers []slog.Handler
}

// Fanout distributes records to multiple slog.Handler in parallel
func Fanout(handlers ...slog.Handler) slog.Handler {
	return &FanoutHandler{
		handlers: handlers,
	}
}

// Implements slog.Handler
func (h *FanoutHandler) Enabled(ctx context.Context, l slog.Level) bool {
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, l) {
			return true
		}
	}

	return false
}

// Implements slog.Handler
// @TODO: return multiple errors ?
func (h *FanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, r.Level) {
			err := h.handlers[i].Handle(ctx, r.Clone())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Implements slog.Handler
func (h *FanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var handlers = make([]slog.Handler, len(h.handlers))
	for i := range handlers {
		handlers[i] = h.handlers[i].WithAttrs(attrs)
	}
	return Fanout(handlers...)
}

// Implements slog.Handler
func (h *FanoutHandler) WithGroup(name string) slog.Handler {
	var handlers = make([]slog.Handler, len(h.handlers))
	for i := range handlers {
		handlers[i] = h.handlers[i].WithGroup(name)
	}
	return Fanout(handlers...)
}
