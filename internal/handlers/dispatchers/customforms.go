package dispatchers

import (
	"log/slog"
	"net/http"

	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/env"
	"github.com/krateoplatformops/snowplow/plumbing/http/response/status"
)

func CustomForm() http.Handler {
	return &customformHandler{
		authnNS: env.String("AUTHN_NAMESPACE", ""),
		verbose: env.True("DEBUG"),
	}
}

type customformHandler struct {
	authnNS string
	verbose bool
}

var _ http.Handler = (*customformHandler)(nil)

func (r *customformHandler) ServeHTTP(wri http.ResponseWriter, req *http.Request) {
	log := xcontext.Logger(req.Context())

	// user logged in check
	if _, err := xcontext.UserConfig(req.Context()); err != nil {
		log.Error("unable to get user endpoint", slog.Any("err", err))
		status.Unauthorized(wri, err)
		return
	}

	// TODO customforms.Resolve(req.Context(), )

}
