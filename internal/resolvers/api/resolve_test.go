//go:build integration
// +build integration

package api

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/krateoplatformops/snowplow/apis"
	v1 "github.com/krateoplatformops/snowplow/apis/templates/v1"
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

func TestMain(m *testing.M) {
	const (
		crdPath      = "../../../crds"
		testdataPath = "../../../testdata"
	)

	xenv.SetTestMode(true)

	namespace = "demo-system"
	clusterName = "krateo"
	testenv = env.New()

	testenv.Setup(
		envfuncs.CreateCluster(kind.NewProvider(), clusterName),
		envfuncs.SetupCRDs(crdPath, "templates.krateo.io_restactions.yaml"),
		e2e.CreateNamespace(namespace),

		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			r, err := resources.New(cfg.Client().RESTConfig())
			if err != nil {
				return ctx, err
			}
			r.WithNamespace(namespace)

			err = decoder.ApplyWithManifestDir(ctx, r, testdataPath, "rbac.yaml", []resources.CreateOption{})
			if err != nil {
				return ctx, err
			}

			time.Sleep(2 * time.Second)
			return ctx, nil
		},
	).Finish(
		envfuncs.DeleteNamespace(namespace),
		envfuncs.TeardownCRDs(crdPath, "templates.krateo.io_restactions.yaml"),
		envfuncs.DestroyCluster(clusterName),
	)

	os.Exit(testenv.Run(m))
}

func TestResolveAPI(t *testing.T) {
	const (
		testdataPath = "../../../testdata"
	)
	os.Setenv("DEBUG", "1")

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
				ctx, os.DirFS(filepath.Join(testdataPath, "restactions")), "*.yaml",
				decoder.CreateHandler(r),
				decoder.MutateNamespace(namespace),
			)
			if err != nil {
				t.Fatal(err)
			}
			return ctx
		}).
		Assess("Resolve API", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			r, err := resources.New(cfg.Client().RESTConfig())
			if err != nil {
				t.Fail()
			}
			r.WithNamespace(namespace)
			apis.AddToScheme(r.GetScheme())

			cr := v1.RESTAction{}
			err = r.Get(ctx, "external-api", namespace, &cr)
			if err != nil {
				t.Fail()
			}

			res := Resolve(ctx, ResolveOptions{
				RC:         cfg.Client().RESTConfig(),
				AuthnNS:    cfg.Namespace(),
				Username:   "cyberjoker",
				UserGroups: []string{"devs"},
				Items:      cr.Spec.API,
			})
			if err != nil {
				t.Fatal(err)
			}

			enc := json.NewEncoder(os.Stderr)
			enc.SetIndent("", "  ")
			if err := enc.Encode(res); err != nil {
				t.Fatal(err)
			}

			return ctx
		}).Feature()

	testenv.Test(t, f)
}
