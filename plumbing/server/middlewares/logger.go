package middlewares

import (
	"net/http"
	"time"

	"log/slog"

	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/shortid"
)

func Logger(root *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(wri http.ResponseWriter, req *http.Request) {
			traceId := req.Header.Get("X-Krateo-TraceId")
			if len(traceId) == 0 {
				traceId = shortid.MustGenerate()
			}

			sub := req.Header.Get("X-Krateo-User")
			orgs := req.Header.Get("X-Krateo-Groups")

			log := root
			if len(sub) > 0 {
				log = root.With("traceId", traceId,
					slog.Group("user",
						slog.String("name", sub),
						slog.String("groups", orgs)),
				)
			}

			ctx := xcontext.BuildContext(req.Context(),
				xcontext.WithTraceId(traceId),
				xcontext.WithLogger(log),
				xcontext.WithRequestStartedAt(time.Now()),
			)

			next.ServeHTTP(wri, req.WithContext(ctx))

			log.Debug("request elapsed time",
				slog.String("duration", xcontext.RequestElapsedTime(req.Context())))
		}

		return http.HandlerFunc(fn)
	}
}
