package use

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/endpoints"
	"github.com/krateoplatformops/snowplow/plumbing/env"
	"github.com/krateoplatformops/snowplow/plumbing/http/response/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
)

func UserConfig() func(http.Handler) http.Handler {
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

			sarc, err := rest.InClusterConfig()
			if err != nil {
				status.InternalError(wri, fmt.Errorf("unable to create in cluster config: %w", err))
				return
			}

			authnNS := env.String("AUTHN_NAMESPACE", "")
			ep, err := endpoints.FromSecret(context.Background(), sarc,
				fmt.Sprintf("%s-clientconfig", sub), authnNS)
			if err != nil {
				if apierrors.IsNotFound(err) {
					status.Unauthorized(wri, err)
					return
				}
				status.InternalError(wri, err)
				return
			}

			// Store the *rest.Config in the context
			ctx := xcontext.BuildContext(req.Context(),
				xcontext.WithUserConfig(ep),
			)

			next.ServeHTTP(wri, req.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}
