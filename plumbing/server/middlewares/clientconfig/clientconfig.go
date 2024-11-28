package clientconfig

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/krateoplatformops/snowplow/plumbing/endpoints"
	"github.com/krateoplatformops/snowplow/plumbing/http/response/status"
	"github.com/krateoplatformops/snowplow/plumbing/kubeconfig"
	"github.com/krateoplatformops/snowplow/plumbing/server"
	"github.com/krateoplatformops/snowplow/plumbing/server/middlewares/logger"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
)

func New(authnNS string, verbose bool) server.Middleware {
	return func(next server.Handler) server.Handler {
		return func(wri http.ResponseWriter, req *http.Request) {
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

			log := logger.Get(req.Context())

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

			rc, err := kubeconfig.NewClientConfig(context.Background(), ep)
			if err != nil {
				log.Error("unable to create user client config", slog.Any("err", err))
				status.InternalError(wri, err)
				return
			}

			// Store the *rest.Config in the context
			ctx := context.WithValue(req.Context(), clientConfigContextKey, rc)

			next(wri, req.WithContext(ctx))
		}
	}
}

// Get retrieves the user *rest.Config from the context.
func Get(ctx context.Context) (*rest.Config, error) {
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
