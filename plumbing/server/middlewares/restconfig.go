package middlewares

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/krateoplatformops/snowplow/plumbing/endpoints"
	"github.com/krateoplatformops/snowplow/plumbing/http/response/status"
	"github.com/krateoplatformops/snowplow/plumbing/kubeconfig"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
)

func RESTConfig(authnNS string, verbose bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(wri http.ResponseWriter, req *http.Request) {
			sub := req.Header.Get("X-Krateo-User")
			orgs := strings.Split(req.Header.Get("X-Krateo-Groups"), ",")

			if len(sub) == 0 {
				status.BadRequest(wri, fmt.Errorf("missing 'X-Krateo-User' header"))
				return
			}

			if len(orgs) == 0 {
				status.BadRequest(wri, fmt.Errorf("missing 'X-Krateo-Groups' header"))
				return
			}

			log := LogFromContext(req.Context())

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

			rc, err := kubeconfig.NewClientConfig(context.Background(), ep)
			if err != nil {
				log.Error("unable to create user client config", slog.Any("err", err))
				status.InternalError(wri, err)
				return
			}

			// Store the *rest.Config in the context
			ctx := context.WithValue(req.Context(), clientConfigContextKey, rc)

			next.ServeHTTP(wri, req.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

// RESTConfigFromContext retrieves the user *rest.Config from the context.
func RESTConfigFromContext(ctx context.Context) (*rest.Config, error) {
	rc, ok := ctx.Value(clientConfigContextKey).(*rest.Config)
	if !ok {
		return nil, fmt.Errorf("user *rest.Config not found in context")
	}

	return rc, nil
}

type contextKey string

const (
	clientConfigContextKey contextKey = "clientconfig"
)
