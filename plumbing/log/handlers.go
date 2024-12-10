package log

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"time"
)

func NewPrettyJSONHandler(output *os.File, opts *slog.HandlerOptions) slog.Handler {
	return &prettyJSONHandler{
		base: slog.NewJSONHandler(output, opts),
	}
}

type prettyJSONHandler struct {
	base slog.Handler
}

func (h *prettyJSONHandler) Handle(ctx context.Context, r slog.Record) error {
	logData := map[string]interface{}{
		"time":  r.Time.Format(time.RFC3339),
		"level": r.Level.String(),
		"msg":   r.Message,
		"pippo": "pluto",
	}

	// Aggiungere gli attributi del log alla mappa
	r.Attrs(func(attr slog.Attr) bool {
		logData[attr.Key] = attr.Value
		return true
	})

	// Convertire la mappa a JSON indentato
	entryBytes, err := json.MarshalIndent(logData, "", "  ")
	if err != nil {
		return err
	}

	// Scrivere l'entry formattata a output
	_, err = os.Stdout.Write(entryBytes)
	if err != nil {
		return err
	}
	_, err = os.Stdout.WriteString("\n")
	return err
}

func (h *prettyJSONHandler) Enabled(ctx context.Context, lev slog.Level) bool {
	return h.base.Enabled(ctx, lev)
}

func (h *prettyJSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h.base.WithAttrs(attrs)
}

func (h *prettyJSONHandler) WithGroup(name string) slog.Handler {
	return h.base.WithGroup(name)
}
