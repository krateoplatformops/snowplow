package context

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/krateoplatformops/snowplow/plumbing/endpoints"
	"github.com/krateoplatformops/snowplow/plumbing/shortid"
)

const (
	LabelKrateoTraceId = "X-Krateo-TraceId"
	LabelKrateoUser    = "X-Krateo-User"
	LabelKrateoGroups  = "X-Krateo-Groups"
)

func Logger(ctx context.Context) *slog.Logger {
	log, ok := ctx.Value(contextKeyLogger).(*slog.Logger)
	if !ok {
		log = slog.New(slog.NewJSONHandler(os.Stderr,
			&slog.HandlerOptions{Level: slog.LevelDebug})).
			With("traceId", TraceId(ctx, false))
	}

	return log
}

func TraceId(ctx context.Context, generate bool) string {
	traceId, ok := ctx.Value(contextKeyTraceId).(string)
	if ok {
		return traceId
	}

	if generate {
		traceId = shortid.MustGenerate()
	}

	return traceId
}

func UserConfig(ctx context.Context) (endpoints.Endpoint, error) {
	ep, ok := ctx.Value(contextKeyUserConfig).(endpoints.Endpoint)
	if !ok {
		return endpoints.Endpoint{}, fmt.Errorf("user *Endpoint not found in context")
	}

	return ep, nil
}

func RequestElapsedTime(ctx context.Context) string {
	start, ok := ctx.Value(contextKeyRequestStartAt).(time.Time)
	if !ok {
		start = time.Now()
	}

	dur := time.Since(start)
	if dur < time.Second {
		return dur.String()
	}

	return dur.Round(time.Microsecond).String()
}

func WithTraceId(traceId string) WithContextFunc {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, contextKeyTraceId, traceId)
	}
}

func WithLogger(log *slog.Logger) WithContextFunc {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, contextKeyLogger, log)
	}
}

func WithRequestStartedAt(t time.Time) WithContextFunc {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, contextKeyRequestStartAt, t)
	}
}

func WithUserConfig(ep endpoints.Endpoint) WithContextFunc {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, contextKeyUserConfig, ep)
	}
}

func BuildContext(ctx context.Context, opts ...WithContextFunc) context.Context {
	for _, fn := range opts {
		ctx = fn(ctx)
	}

	return ctx
}

type WithContextFunc func(context.Context) context.Context

type contextKey string

func (c contextKey) String() string {
	return "snowplow." + string(c)
}

var (
	contextKeyTraceId = contextKey("traceId")
	contextKeyLogger  = contextKey("logger")
	//contextKeyRESTConfig     = contextKey("restConfig")
	contextKeyUserConfig     = contextKey("userConfig")
	contextKeyRequestStartAt = contextKey("requestStartedAt")
)
