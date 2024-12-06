//go:build integration
// +build integration

package actions

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/krateoplatformops/snowplow/apis"
	"github.com/krateoplatformops/snowplow/apis/templates/v1alpha1"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/e2e"

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
		crdsPath = "../../../crds"
	)

	namespace = "test-system"
	clusterName = envconf.RandomName("krateo", 16)
	testenv = env.New()

	testenv.Setup(
		envfuncs.CreateCluster(kind.NewProvider(), clusterName),
		envfuncs.SetupCRDs(crdsPath, "templates.krateo.io_customforms.yaml"),
		envfuncs.CreateNamespace(namespace),

		func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
			// TODO: add a wait.For conditional helper that can
			// check and wait for the existence of a CRD resource
			time.Sleep(2 * time.Second)
			return ctx, nil
		},
	).Finish(
		envfuncs.DeleteNamespace(namespace),
		envfuncs.TeardownCRDs(crdsPath, "*"),
		envfuncs.DestroyCluster(clusterName),
	)

	os.Exit(testenv.Run(m))
}

func TestCustomFormActions(t *testing.T) {
	const (
		testdata = "../../../testdata/customforms"
	)

	f := features.New("custom form").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			r, err := resources.New(c.Client().RESTConfig())
			if err != nil {
				t.Fail()
			}

			apis.AddToScheme(r.GetScheme())

			r.WithNamespace(namespace)

			err = decoder.DecodeEachFile(
				ctx, os.DirFS(testdata), "*",
				decoder.CreateHandler(r),
				decoder.MutateNamespace(namespace),
			)
			if err != nil {
				t.Fatal(err)
			}
			return ctx
		}).
		Assess("Check If Resource created", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			r, err := resources.New(c.Client().RESTConfig())
			if err != nil {
				t.Fail()
			}
			r.WithNamespace(namespace)
			apis.AddToScheme(r.GetScheme())

			cf := &v1alpha1.CustomForm{}
			err = r.Get(ctx, "fireworksapp", namespace, cf)
			if err != nil {
				t.Fail()
			}

			log := xcontext.Logger(ctx)
			log.Info("CR Details", slog.Any("cr", cf))

			return ctx
		}).Feature()

	testenv.
		Setup(e2e.Logger(), e2e.SignUp("cyberjoker", []string{"devs"})).
		Test(t, f)
}

func TestResolveActions(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(os.Stdout,
		&slog.HandlerOptions{Level: slog.LevelDebug}))

	ctx := xcontext.BuildContext(context.TODO(),
		xcontext.WithTraceId("test"),
		xcontext.WithLogger(log),
	)

	res, err := Resolve(ctx, []*v1alpha1.Action{
		{
			Template: &v1alpha1.ActionTemplate{
				ID:         "test-id",
				Name:       "nginx",
				Namespace:  "demo-system",
				Resource:   "deployments",
				APIVersion: "apps/v1",
				Verb:       "put",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(res)
}
