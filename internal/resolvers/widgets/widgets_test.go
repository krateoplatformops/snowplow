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
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/e2e"
	xenv "github.com/krateoplatformops/snowplow/plumbing/env"
	"github.com/krateoplatformops/snowplow/plumbing/kubeutil"

	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/support/kind"
)

var (
	testenv     env.Environment
	clusterName string
	namespace   string
)

const (
	crdPath       = "../../../crds"
	manifestsPath = "../../../testdata"
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

		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			r, err := resources.New(cfg.Client().RESTConfig())
			if err != nil {
				return ctx, err
			}
			r.WithNamespace(namespace)

			err = decoder.ApplyWithManifestDir(ctx, r, manifestsPath, "rbac.yaml", []resources.CreateOption{})
			if err != nil {
				return ctx, err
			}

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

func TestWidget(t *testing.T) {
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

			apis.AddToScheme(r.GetScheme())

			r.WithNamespace(namespace)

			err = decoder.DecodeEachFile(
				ctx, os.DirFS(filepath.Join(manifestsPath, "widgets")), "*.yaml",
				decoder.CreateHandler(r),
				decoder.MutateNamespace(namespace),
			)
			if err != nil {
				t.Fatal(err)
			}
			return ctx
		}).
		Assess("Resolve Widget", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			r, err := resources.New(c.Client().RESTConfig())
			if err != nil {
				t.Fail()
			}
			r.WithNamespace(namespace)
			apis.AddToScheme(r.GetScheme())

			cr := v1alpha1.Widget{}
			err = r.Get(ctx, "external-api", namespace, &cr)
			if err != nil {
				t.Fail()
			}

			res, err := Resolve(ctx, ResolveOptions{
				In:         &cr,
				SArc:       c.Client().RESTConfig(),
				AuthnNS:    namespace,
				Username:   "cyberjoker",
				UserGroups: []string{"devs"},
			})
			if err != nil {
				log := xcontext.Logger(ctx)
				log.Error("unable to resolve widget", slog.Any("err", err))
				t.Fail()
			}

			res.Kind = "Widget"
			res.APIVersion = v1alpha1.SchemeGroupVersion.String()
			if err := kubeutil.ToYAML(os.Stderr, res); err != nil {
				t.Fatal(err)
			}

			return ctx
		}).Feature()

	testenv.Test(t, f)
}
