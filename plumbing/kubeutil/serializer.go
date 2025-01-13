package kubeutil

import (
	"io"

	"github.com/krateoplatformops/snowplow/apis"
	"k8s.io/apimachinery/pkg/runtime"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

func ToYAML(wri io.Writer, obj runtime.Object) error {
	rs := runtime.NewScheme()
	if err := apis.AddToScheme(rs); err != nil {
		return err
	}

	s := serializer.NewSerializerWithOptions(serializer.DefaultMetaFactory, rs, rs,
		serializer.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: false,
		})

	return s.Encode(obj, wri)
}
