//go:build integration
// +build integration

package apiref

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/krateoplatformops/snowplow/apis"
	"github.com/krateoplatformops/snowplow/internal/objects"
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

var (
	testenv     env.Environment
	clusterName string
	namespace   string
)

const (
	crdPath      = "../../../../crds"
	testdataPath = "../../../../testdata"
)

func TestMain(m *testing.M) {
	xenv.SetTestMode(true)

	namespace = "demo-system"
	clusterName = "krateo"
	testenv = env.New()

	testenv.Setup(
		envfuncs.CreateCluster(kind.NewProvider(), clusterName),
		envfuncs.SetupCRDs(crdPath, "templates.krateo.io_restactions.yaml"),
		envfuncs.SetupCRDs(filepath.Join(testdataPath, "widgets"), "widgets.templates.krateo.io_buttons.yaml"),
		e2e.CreateNamespace(namespace),

		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			r, err := resources.New(cfg.Client().RESTConfig())
			if err != nil {
				return ctx, err
			}
			r.WithNamespace(namespace)

			err = decoder.ApplyWithManifestDir(ctx, r, testdataPath, "rbac.widgets.yaml", []resources.CreateOption{})
			if err != nil {
				return ctx, err
			}

			err = decoder.ApplyWithManifestDir(ctx, r, testdataPath, "rbac.restactions.yaml", []resources.CreateOption{})
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
		envfuncs.TeardownCRDs(filepath.Join(testdataPath, "widgets"), "widgets.templates.krateo.io_buttons.yaml"),
		envfuncs.DestroyCluster(clusterName),
		e2e.Coverage(),
	)

	os.Exit(testenv.Run(m))
}

func TestResolveWidgets(t *testing.T) {
	os.Setenv("DEBUG", "0")

	f := features.New("Setup").
		Setup(e2e.Logger("test")).
		Setup(e2e.SignUp("cyberjoker", []string{"devs"}, namespace)).
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			r, err := resources.New(cfg.Client().RESTConfig())
			if err != nil {
				t.Fail()
			}

			apis.AddToScheme(r.GetScheme())

			r.WithNamespace(namespace)

			err = decoder.DecodeEachFile(
				ctx, os.DirFS(filepath.Join(testdataPath, "widgets")), "button.*.yaml",
				decoder.CreateHandler(r),
				decoder.MutateNamespace(namespace),
			)
			if err != nil {
				t.Fatal(err)
			}
			return ctx
		}).
		Assess("Resolve Widget NO API Reference", resolveApiRefFromWidget("button-sample")).
		Assess("Resolve Widget API Reference", resolveApiRefFromWidget("button-with-api")).
		Feature()

	testenv.Test(t, f)
}

func resolveApiRefFromWidget(name string) func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {

		res := objects.Get(ctx, objects.Reference{
			Name: name, Namespace: namespace,
			Resource: "buttons", APIVersion: "widgets.templates.krateo.io/v1beta1",
		})
		if res.Err != nil {
			log := xcontext.Logger(ctx)
			log.Error("unable to get object", slog.Any("err", res.Err))
			t.Fail()
		}

		obj, err := Resolve(ctx, ResolveOptions{
			RC:     c.Client().RESTConfig(),
			Widget: res.Unstructured,
		})
		if err != nil {
			log := xcontext.Logger(ctx)
			log.Error("unable to resolve api reference", slog.Any("err", err))
			t.Fail()
		}

		enc := json.NewEncoder(os.Stderr)
		enc.SetIndent(" ", "   ")
		if err := enc.Encode(obj); err != nil {
			t.Fatal(err)
		}

		return ctx
	}
}
