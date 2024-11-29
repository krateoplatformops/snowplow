package middlewares

import (
	"context"
	"net/http"
	"os"
	"time"

	"log/slog"

	"github.com/krateoplatformops/snowplow/plumbing/server/traceid"
	"github.com/krateoplatformops/snowplow/plumbing/shortid"
)

func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(wri http.ResponseWriter, req *http.Request) {

			tid := req.Header.Get("X-Request-Id")
			if len(tid) == 0 {
				tid = shortid.MustGenerate()
			}

			sub := req.Header.Get("X-Krateo-User")
			orgs := req.Header.Get("X-Krateo-Groups")

			ctx := traceid.Set(req.Context())
			ctx = context.WithValue(ctx, startTimeKey, time.Now())

			if len(sub) == 0 {
				ctx = context.WithValue(ctx, logKey, log.With("traceId", tid))
			} else {
				ctx = context.WithValue(ctx, logKey,
					log.With("traceId", tid,
						slog.Group("user",
							slog.String("name", sub),
							slog.String("groups", orgs)),
					))
			}

			next.ServeHTTP(wri, req.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

// LogFromContext retrieves the logger from the request context.
func LogFromContext(ctx context.Context) *slog.Logger {
	log, ok := ctx.Value(logKey).(*slog.Logger)
	if !ok {
		log = slog.New(slog.NewJSONHandler(os.Stderr,
			&slog.HandlerOptions{Level: slog.LevelDebug})).
			With("traceId", traceid.Get(ctx))
	}

	return log
}

func ElapsedTime(ctx context.Context) string {
	start, ok := ctx.Value(startTimeKey).(time.Time)
	if !ok {
		start = time.Now()
	}

	return time.Since(start).Round(time.Microsecond).String()
}

const (
	logKey       contextKey = "x-request-log"
	startTimeKey contextKey = "x-request-start"
)