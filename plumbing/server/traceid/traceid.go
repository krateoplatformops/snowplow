package traceid

import (
	"context"

	"github.com/krateoplatformops/snowplow/plumbing/shortid"
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

	traceID = shortid.MustGenerate()
	return context.WithValue(ctx, requestIdContextKey, traceID)
}

type contextKey string

const (
	requestIdContextKey contextKey = "X-Request-Id"
)
