package context

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/krateoplatformops/snowplow/plumbing/endpoints"
	"github.com/krateoplatformops/snowplow/plumbing/env"
	"github.com/krateoplatformops/snowplow/plumbing/shortid"
	"github.com/krateoplatformops/snowplow/plumbing/tmpl"
)

const (
	Trace uint = 1 << iota // 1 << 0 = 1
	Debug                  // 1 << 1 = 2
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
	if !env.TestMode() {
		ep.ServerURL = "https://kubernetes.default.svc"
	}
	return ep, nil
}

func JQTemplate(ctx context.Context) tmpl.JQTemplate {
	v := ctx.Value(contextKeyJQTemplate)
	if val, ok := v.(tmpl.JQTemplate); ok {
		return val
	}
	return nil
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

func WithLogger(root *slog.Logger) WithContextFunc {
	return func(ctx context.Context) context.Context {
		if root == nil {
			logLevel := slog.LevelInfo
			if env.True("DEBUG") {
				logLevel = slog.LevelDebug
			}
			root = slog.New(slog.NewJSONHandler(os.Stderr,
				&slog.HandlerOptions{Level: logLevel}))
		}

		return context.WithValue(ctx, contextKeyLogger,
			root.With("traceId", TraceId(ctx, false)))
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

func WithJQTemplate() WithContextFunc {
	return func(ctx context.Context) context.Context {
		tpl, err := tmpl.New("${", "}")
		if err != nil {
			Logger(ctx).Error("unable to create jq template engine", slog.Any("err", err))
			return ctx
		}

		return context.WithValue(ctx, contextKeyJQTemplate, tpl)
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
	contextKeyUserConfig     = contextKey("userConfig")
	contextKeyRequestStartAt = contextKey("requestStartedAt")
	contextKeyJQTemplate     = contextKey("jqTemplateEngine")
	contextKeyAuthnNS        = contextKey("authnNS")
)
