package dynamic

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	cacheddiscovery "k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

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
