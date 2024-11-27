package traceid

import (
	"context"

	"github.com/google/uuid"
)

func Get(ctx context.Context) string {
	id, _ := ctx.Value(requestIdContextKey).(string)
	return id
}

func Set(ctx context.Context) context.Context {
	traceID, ok := ctx.Value(requestIdContextKey).(string)
	if ok && traceID != "" {
		return ctx
	}

	traceID = uuidV7()
	return context.WithValue(ctx, requestIdContextKey, traceID)
}

func uuidV7() string {
	return uuid.Must(uuid.NewV7()).String()
}

type contextKey string

const (
	requestIdContextKey contextKey = "X-Request-Id"
)
