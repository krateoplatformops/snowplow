//go:build integration
// +build integration

package definitions

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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
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
	clusterName = "krateo" //envconf.RandomName("krateo", 16)
	testenv = env.New()

	testenv.Setup(
		envfuncs.CreateCluster(kind.NewProvider(), clusterName),
		envfuncs.SetupCRDs(crdPath, "templates.krateo.io_forms.yaml"),
		envfuncs.SetupCRDs(filepath.Join(testdataPath, "/forms"), "core.krateo.io_compositiondefinitions.yaml"),
		e2e.CreateNamespace(namespace),

		func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
			// TODO: add a wait.For conditional helper that can
			// check and wait for the existence of a CRD resource
			time.Sleep(2 * time.Second)
			return ctx, nil
		},
	).Finish(
		envfuncs.DeleteNamespace(namespace),
		envfuncs.TeardownCRDs(crdPath, "templates.krateo.io_forms.yaml"),
		envfuncs.TeardownCRDs(filepath.Join(testdataPath, "/forms"), "core.krateo.io_compositiondefinitions.yaml"),
		envfuncs.DestroyCluster(clusterName),
	)

	os.Exit(testenv.Run(m))
}

func TestResolveDefinition(t *testing.T) {
	const (
		testdataPath = "../../../testdata"
	)

	os.Setenv("DEBUG", "false")

	patchData, err := os.ReadFile(filepath.Join(testdataPath, "forms", "fireworksapp-patch.json"))
	if err != nil {
		t.Fail()
	}

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

			fin, err := os.Open(filepath.Join(testdataPath, "forms", "fireworksapp.yaml"))
			if err != nil {
				t.Fatal(err)
			}
			defer fin.Close()

			obj, err := decoder.DecodeAny(fin, decoder.MutateNamespace(namespace))
			if err != nil {
				t.Fatal(err)
			}

			err = r.Create(ctx, obj)
			if err != nil {
				t.Fatal(err)
			}
			//spew.Dump(obj)
			err = r.PatchStatus(ctx, obj, k8s.Patch{
				PatchType: types.MergePatchType,
				Data:      patchData,
			})
			if err != nil {
				t.Fatal(err)
			}

			err = decoder.DecodeEachFile(
				ctx, os.DirFS(filepath.Join(testdataPath, "forms")), "rbac.yaml",
				decoder.CreateHandler(r),
				decoder.MutateNamespace(namespace),
			)
			if err != nil {
				if !apierrors.IsAlreadyExists(err) {
					t.Fatal(err)
				}
			}

			err = decoder.DecodeEachFile(
				ctx, os.DirFS(filepath.Join(testdataPath, "forms")), "sample.yaml",
				decoder.CreateHandler(r),
				decoder.MutateNamespace(namespace),
			)
			if err != nil {
				if !apierrors.IsAlreadyExists(err) {
					t.Fatal(err)
				}
			}

			r.Get(ctx, "", namespace, obj)

			return ctx
		}).
		Assess("Resolve form", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			r, err := resources.New(c.Client().RESTConfig())
			if err != nil {
				t.Fail()
			}
			r.WithNamespace(namespace)
			apis.AddToScheme(r.GetScheme())

			cr := v1alpha1.Form{}
			err = r.Get(ctx, "fireworksapp", namespace, &cr)
			if err != nil {
				t.Fail()
			}

			log := xcontext.Logger(ctx)
			log.Debug("form fetched", slog.Any("cr", cr))

			res, err := Resolve(ctx, &cr)
			if err != nil {
				log.Error("unable to resolve form", slog.Any("err", err))
				t.Fail()
			} else {
				log.Info("resolved form", slog.Any("result", res))
			}

			return ctx
		}).Feature()

	testenv.Test(t, f)
}
