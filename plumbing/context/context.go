package context

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/krateoplatformops/snowplow/plumbing/shortid"
	"k8s.io/client-go/rest"
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

// RESTConfig retrieves the user *rest.Config from the context.
func RESTConfig(ctx context.Context) (*rest.Config, error) {
	rc, ok := ctx.Value(contextKeyRESTConfig).(*rest.Config)
	if !ok {
		return nil, fmt.Errorf("user *rest.Config not found in context")
	}

	return rc, nil
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

func WithRESTConfig(rc *rest.Config) WithContextFunc {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, contextKeyRESTConfig, rc)
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
	contextKeyTraceId        = contextKey("traceId")
	contextKeyLogger         = contextKey("logger")
	contextKeyRESTConfig     = contextKey("restConfig")
	contextKeyRequestStartAt = contextKey("requestStartedAt")
)
