package e2e

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/env"
	"github.com/krateoplatformops/snowplow/plumbing/kubeconfig"
	"github.com/krateoplatformops/snowplow/plumbing/log"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/types"
	"sigs.k8s.io/yaml"
)

func JQTemplate() types.StepFunc {
	return func(ctx context.Context, _ *testing.T, _ *envconf.Config) context.Context {
		return xcontext.BuildContext(ctx,
			xcontext.WithJQ(),
		)
	}
}

func Logger(traceId string) types.StepFunc {
	logLevel := slog.LevelInfo
	if env.True("DEBUG") {
		logLevel = slog.LevelDebug
	}
	handler := log.NewPrettyJSONHandler(os.Stderr,
		&slog.HandlerOptions{Level: logLevel})

	return func(ctx context.Context, _ *testing.T, _ *envconf.Config) context.Context {
		return xcontext.BuildContext(ctx,
			xcontext.WithTraceId(traceId),
			xcontext.WithLogger(slog.New(handler)),
		)
	}
}

func SignUp(user string, groups []string, namespace string) types.StepFunc {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		dat, err := os.ReadFile(cfg.KubeconfigFile())
		if err != nil {
			t.Fatal(err)
		}

		in := kubeconfig.KubeConfig{}
		if err := yaml.Unmarshal(dat, &in); err != nil {
			t.Fatal(err)
		}

		handler := &signupHandler{
			restconfig:   cfg.Client().RESTConfig(),
			namespace:    namespace,
			caData:       in.Clusters[0].Cluster.CertificateAuthorityData,
			serverURL:    in.Clusters[0].Cluster.Server, //"https://kubernetes.default.svc",
			certDuration: time.Minute * 30,
		}

		ep, err := handler.SignUp(user, groups)
		if err != nil {
			t.Fatal(err)
		}

		return xcontext.BuildContext(ctx, xcontext.WithUserConfig(ep))
	}
}
