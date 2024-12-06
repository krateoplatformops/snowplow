package e2e

import (
	"context"
	"log/slog"
	"os"
	"time"

	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/kubeconfig"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/yaml"
)

func Logger() env.Func {
	return func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		log := slog.New(slog.NewJSONHandler(os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug})).
			With("traceId", xcontext.TraceId(ctx, true))

		return xcontext.BuildContext(ctx,
			xcontext.WithLogger(log),
		), nil
	}
}

func SignUp(user string, groups []string) env.Func {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		dat, err := os.ReadFile(cfg.KubeconfigFile())
		if err != nil {
			return ctx, err
		}

		in := kubeconfig.KubeConfig{}
		if err := yaml.Unmarshal(dat, &in); err != nil {
			return ctx, err
		}

		handler := &signupHandler{
			restconfig:   cfg.Client().RESTConfig(),
			namespace:    cfg.Namespace(),
			caData:       in.Clusters[0].Cluster.CertificateAuthorityData,
			serverURL:    "https://kubernetes.default.svc",
			certDuration: time.Minute * 30,
		}

		ep, err := handler.SignUp(user, groups)
		if err != nil {
			return ctx, err
		}

		return xcontext.BuildContext(ctx, xcontext.WithUserConfig(ep)), nil
	}
}
