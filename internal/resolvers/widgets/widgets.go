package widgets

import (
	"context"
	"fmt"
	"net/http"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeschema "k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/krateoplatformops/snowplow/internal/dynamic"
	crdschema "github.com/krateoplatformops/snowplow/internal/resolvers/crds/schema"
	"github.com/krateoplatformops/snowplow/plumbing/env"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

const (
	widgetDataKey = "widgetData"
	apiKey        = "api"
)

type Widget = unstructured.Unstructured

type ResolveOptions struct {
	In         *Widget
	RC         *rest.Config
	AuthnNS    string
	Username   string
	UserGroups []string
}

func Resolve(ctx context.Context, opts ResolveOptions) (*Widget, error) {
	spec, err := nestedSpecCopy(opts.In.Object, widgetDataKey)
	if err != nil {
		return opts.In, err
	}

	err = unstructured.SetNestedMap(opts.In.Object, spec, "status")
	if err != nil {
		return opts.In, err
	}

	if env.TestMode() {
		err = crdschema.ValidateObjectStatus(ctx, opts.RC, opts.In.Object)
	} else {
		err = crdschema.ValidateObjectStatus(ctx, nil, opts.In.Object)
	}
	if err != nil {
		return opts.In, err
	}

	return opts.In, nil
}

func nestedSpecCopy(obj map[string]any, key string) (map[string]any, error) {
	data, ok, err := unstructured.NestedMap(obj, "spec", key)
	if err != nil {
		return map[string]any{}, err
	}
	if !ok {
		name := dynamic.GetName(obj)
		namespace := dynamic.GetNamespace(obj)
		gv, _ := runtimeschema.ParseGroupVersion(dynamic.GetAPIVersion(obj))
		err := &apierrors.StatusError{
			ErrStatus: metav1.Status{
				Status: metav1.StatusFailure,
				Code:   http.StatusNotFound,
				Reason: metav1.StatusReasonNotFound,
				Details: &metav1.StatusDetails{
					Group: gv.Group,
					Kind:  dynamic.GetKind(obj),
					Name:  name,
				},
				Message: fmt.Sprintf("spec %q not found in %s @ %s", key, name, namespace),
			}}
		return map[string]any{}, err
	}

	return runtime.DeepCopyJSON(data), nil
}

/*
func resolveRESTActionRef(ctx context.Context, obj map[string]any) error {
	api, ok, err := unstructured.NestedMap(obj, "spec", apiKey)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	var ra v1.RESTAction
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(api, &ra)
	if err != nil {
		return err
	}

	_, err = restactions.Resolve(ctx, restactions.ResolveOptions{
		In:         &ra,
		AuthnNS:    r.authnNS,
		Username:   r.username,
		UserGroups: r.userGroups,
	})
	if err != nil {
		return err
	}

	return nil
}
*/
