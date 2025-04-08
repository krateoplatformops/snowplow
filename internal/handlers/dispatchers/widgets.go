package dispatchers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/krateoplatformops/snowplow/internal/resolvers/widgets"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/env"
	"github.com/krateoplatformops/snowplow/plumbing/http/response"
)

func Widgets() http.Handler {
	return &widgetsHandler{
		authnNS: env.String("AUTHN_NAMESPACE", ""),
		verbose: env.True("DEBUG"),
	}
}

type widgetsHandler struct {
	authnNS string
	verbose bool
}

var _ http.Handler = (*widgetsHandler)(nil)

func (r *widgetsHandler) ServeHTTP(wri http.ResponseWriter, req *http.Request) {
	log := xcontext.Logger(req.Context())

	got := fetchObject(req)
	if got.Err != nil {
		response.Encode(wri, got.Err)
		return
	}
	spew.Dump(got)

	ctx := xcontext.BuildContext(req.Context())

	res, err := widgets.Resolve(ctx, widgets.ResolveOptions{
		In:         got.Unstructured,
		Username:   req.Header.Get(xcontext.LabelKrateoUser),
		UserGroups: strings.Split(req.Header.Get(xcontext.LabelKrateoGroups), ","),
		AuthnNS:    r.authnNS,
	})
	if err != nil {
		log.Error("unable to resolve widget",
			slog.String("name", got.Unstructured.GetName()),
			slog.String("namespace", got.Unstructured.GetNamespace()),
			slog.Any("err", err))
		response.InternalError(wri, err)
		return
	}

	wri.Header().Set("Content-Type", "application/json")
	wri.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(wri)
	enc.SetIndent("", "  ")
	enc.Encode(res)
}
