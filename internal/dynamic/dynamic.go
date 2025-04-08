package dynamic

import (
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	xenv "github.com/krateoplatformops/snowplow/plumbing/env"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	cacheddiscovery "k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

func ResourceFor(rc *rest.Config, gvk schema.GroupVersionKind) (schema.GroupVersionResource, error) {
	if rc == nil && !xenv.TestMode() {
		var err error
		rc, err = rest.InClusterConfig()
		if err != nil {
			return schema.GroupVersionResource{}, err
		}
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(rc)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(
		memory.NewMemCacheClient(discoveryClient),
	)

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	return mapping.Resource, nil
}

func KindFor(rc *rest.Config, gvr schema.GroupVersionResource) (gvk schema.GroupVersionKind, err error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(rc)
	if err != nil {
		return gvk, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(
		cacheddiscovery.NewMemCacheClient(discoveryClient),
	)

	gvk, err = mapper.KindFor(gvr)

	return
}

func GroupVersion(obj map[string]any) schema.GroupVersion {
	av := getNestedString(obj, "apiVersion")

	if (len(av) == 0) || (av == "/") {
		return schema.GroupVersion{}
	}

	switch strings.Count(av, "/") {
	case 0:
		return schema.GroupVersion{"", av}
	case 1:
		i := strings.Index(av, "/")
		return schema.GroupVersion{av[:i], av[i+1:]}
	default:
		return schema.GroupVersion{}
	}
}

func GetAPIVersion(obj map[string]any) string {
	return getNestedString(obj, "apiVersion")
}

func GetKind(obj map[string]any) string {
	return getNestedString(obj, "kind")
}

func GetNamespace(obj map[string]any) string {
	return getNestedString(obj, "metadata", "namespace")
}

func GetName(obj map[string]any) string {
	return getNestedString(obj, "metadata", "name")
}

func GetUID(obj map[string]any) types.UID {
	return types.UID(getNestedString(obj, "metadata", "uid"))
}

func getNestedString(obj map[string]any, fields ...string) string {
	val, found, err := unstructured.NestedString(obj, fields...)
	if !found || err != nil {
		return ""
	}
	return val
}
