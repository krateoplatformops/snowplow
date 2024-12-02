package middlewares

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/endpoints"
	"github.com/krateoplatformops/snowplow/plumbing/http/response/status"
	"github.com/krateoplatformops/snowplow/plumbing/kubeconfig"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
)

func RESTConfig(authnNS string, verbose bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(wri http.ResponseWriter, req *http.Request) {
			sub := req.Header.Get(xcontext.LabelKrateoUser)
			orgs := strings.Split(req.Header.Get(xcontext.LabelKrateoGroups), ",")

			if len(sub) == 0 {
				status.BadRequest(wri, fmt.Errorf("missing '%s' header", xcontext.LabelKrateoUser))
				return
			}

			if len(orgs) == 0 {
				status.BadRequest(wri, fmt.Errorf("missing '%s' header", xcontext.LabelKrateoGroups))
				return
			}

			log := xcontext.Logger(req.Context())

			sarc, err := rest.InClusterConfig()
			if err != nil {
				status.InternalError(wri, fmt.Errorf("unable to create in cluster config: %w", err))
				return
			}

			ep, err := endpoints.FromSecret(context.Background(), sarc, fmt.Sprintf("%s-clientconfig", sub), authnNS)
			if err != nil {
				if apierrors.IsNotFound(err) {
					status.Unauthorized(wri, err)
					return
				}
				status.InternalError(wri, err)
				return
			}
			if verbose {
				ep.Debug = true
			}

			rc, err := kubeconfig.NewClientConfig(req.Context(), ep)
			if err != nil {
				log.Error("unable to create user client config", slog.Any("err", err))
				status.InternalError(wri, err)
				return
			}

			// Store the *rest.Config in the context
			ctx := xcontext.BuildContext(req.Context(),
				xcontext.WithRESTConfig(rc),
			)

			next.ServeHTTP(wri, req.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}
