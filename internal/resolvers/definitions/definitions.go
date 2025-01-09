package definitions

import (
	"context"
	"fmt"
	"strings"

	"log/slog"

	"github.com/gobuffalo/flect"
	templates "github.com/krateoplatformops/snowplow/apis/templates/v1alpha1"
	"github.com/krateoplatformops/snowplow/internal/dynamic"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/kubeconfig"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	schemaDefinitionGVK      = schema.FromAPIVersionAndKind("core.krateo.io/v1alpha1", "SchemaDefinition")
	compositionDefinitionGVK = schema.FromAPIVersionAndKind("core.krateo.io/v1alpha1", "CompositionDefinition")
)

type Definition struct {
	GVR       schema.GroupVersionResource
	Kind      string
	Name      string
	Namespace string
}

func Resolve(ctx context.Context, in *templates.Form) (def Definition, err error) {
	log := xcontext.Logger(ctx)

	gvk := schemaDefinitionGVK
	ref := in.Spec.SchemaDefinitionRef
	if ref == nil {
		gvk = compositionDefinitionGVK
		ref = in.Spec.CompositionDefinitionRef
	}

	if ref == nil {
		log.Error("both 'schemaDefinitionRef' and 'compositionDefinitionRef' are undefined",
			slog.String("name", in.Name), slog.String("namespace", in.Namespace))
		return
	}

	ep, err := xcontext.UserConfig(ctx)
	if err != nil {
		log.Error("unable to get user endpoint", slog.Any("err", err))
		return def, err
	}

	rc, err := kubeconfig.NewClientConfig(ctx, ep)
	if err != nil {
		log.Error("unable to create user client config", slog.Any("err", err))
		return def, err
	}

	dyn, err := dynamic.NewClient(rc)
	if err != nil {
		return def, err
	}

	uns, err := dyn.Get(ctx, ref.Name, dynamic.Options{
		Namespace: ref.Namespace,
		GVK:       gvk,
	})
	if err != nil {
		return def, err
	}

	status, ok, err := unstructured.NestedMap(uns.UnstructuredContent(), "status")
	if err != nil {
		return def, err
	}
	if !ok {
		err = fmt.Errorf("status not found in '%s %s/%s'", gvk.String(), ref.Namespace, ref.Name)
		return
	}

	apiVersion, ok := status["apiVersion"].(string)
	if !ok {
		err = fmt.Errorf("status.apiVersion not found in '%s %s/%s'", gvk.String(), ref.Namespace, ref.Name)
		return
	}

	def.Kind, ok = status["kind"].(string)
	if !ok {
		err = fmt.Errorf("status.kind not found in '%s %s/%s'", gvk.String(), ref.Namespace, ref.Name)
		return
	}
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return
	}

	def.GVR = gv.WithResource(flect.Pluralize(strings.ToLower(def.Kind)))
	def.Name = ref.Name
	def.Namespace = ref.Namespace

	return
}
