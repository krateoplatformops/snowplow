//go:build integration
// +build integration

package widgets

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/krateoplatformops/snowplow/apis"
	"github.com/krateoplatformops/snowplow/apis/templates/v1alpha1"
	apiresolver "github.com/krateoplatformops/snowplow/internal/resolvers/api"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/e2e"
	xenv "github.com/krateoplatformops/snowplow/plumbing/env"

	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/support/kind"
)

const (
	crdPath      = "../../../../crds"
	testdataPath = "../../../../testdata"
)

var (
	testenv     env.Environment
	clusterName string
	namespace   string
)

func TestMain(m *testing.M) {
	xenv.SetTestMode(true)

	namespace = "demo-system"
	clusterName = "krateo"
	testenv = env.New()

	testenv.Setup(
		envfuncs.CreateCluster(kind.NewProvider(), clusterName),
		envfuncs.SetupCRDs(crdPath, "templates.krateo.io_widgets.yaml"),
		e2e.CreateNamespace(namespace),

		func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
			// TODO: add a wait.For conditional helper that can
			// check and wait for the existence of a CRD resource
			time.Sleep(2 * time.Second)
			return ctx, nil
		},
	).Finish(
		envfuncs.DeleteNamespace(namespace),
		envfuncs.TeardownCRDs(crdPath, "templates.krateo.io_widgets.yaml"),
		envfuncs.DestroyCluster(clusterName),
	)

	os.Exit(testenv.Run(m))
}

func TestCustomFormApp(t *testing.T) {
	os.Setenv("DEBUG", "false")

	f := features.New("Setup").
		Setup(e2e.Logger("test")).
		Setup(e2e.JQTemplate()).
		Setup(e2e.SignUp("cyberjoker", []string{"devs"}, namespace)).
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {

			r, err := resources.New(cfg.Client().RESTConfig())
			if err != nil {
				t.Fail()
			}
			r.WithNamespace(namespace)

			apis.AddToScheme(r.GetScheme())

			err = decoder.DecodeEachFile(
				ctx, os.DirFS(filepath.Join(testdataPath, "widgets")), "*.yaml",
				decoder.CreateHandler(r),
				decoder.MutateNamespace(namespace),
			)
			if err != nil {
				t.Fatal(err)
			}
			return ctx
		}).
		Assess("Resolve Widget->spec.app", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			log := xcontext.Logger(ctx)

			r, err := resources.New(cfg.Client().RESTConfig())
			if err != nil {
				log.Error(err.Error())
				t.Fail()
			}
			r.WithNamespace(namespace)
			apis.AddToScheme(r.GetScheme())

			cr := v1alpha1.Widget{}
			err = r.Get(ctx, "external-api", namespace, &cr)
			if err != nil {
				t.Fatal(err)
			}

			//log.Debug("widget manifest", slog.Any("cr", cr))

			dict, err := apiresolver.Resolve(ctx, cr.Spec.API, apiresolver.ResolveOptions{
				SARc:       cfg.Client().RESTConfig(),
				AuthnNS:    namespace,
				Username:   "cyberjoker",
				UserGroups: []string{"devs"},
			})
			if err != nil {
				log.Error("unable to resolve api", slog.Any("err", err))
				t.Fail()
			}

			res := Resolve(ctx, cr.Spec.App, dict)
			if err != nil {
				log.Error("unable to resolve app template", slog.Any("err", err))
				t.Fail()
			} else {
				log.Info("App in Status", slog.Any("app", res))
			}

			log.Info("Resolved Widget spec.app",
				slog.String("name", cr.Name), slog.String("namespace", cr.Namespace),
				slog.Any("app", res))

			return ctx
		}).Feature()

	testenv.Test(t, f)
}
