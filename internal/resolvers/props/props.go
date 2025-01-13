package props

import (
	"context"
	"log/slog"

	templates "github.com/krateoplatformops/snowplow/apis/templates/v1alpha1"
	"github.com/krateoplatformops/snowplow/internal/objects"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func Resolve(ctx context.Context, ref *templates.Reference) map[string]string {
	if ref == nil {
		return map[string]string{}
	}

	log := xcontext.Logger(ctx)

	got := objects.Get(ctx, objects.Reference{
		Name: ref.Name, Namespace: ref.Namespace,
		APIVersion: "v1", Resource: "configmaps",
	})
	if got.Err != nil {
		log.Error(got.Err.Message, slog.Any("reference", ref), slog.Any("err", got.Err.Message))
		return map[string]string{}
	}

	sb := runtime.NewSchemeBuilder(
		func(reg *runtime.Scheme) error {
			reg.AddKnownTypes(
				schema.GroupVersion{Version: "v1"},
				&corev1.ConfigMap{},
				&corev1.ConfigMapList{},
				&metav1.ListOptions{},
				&metav1.GetOptions{},
				&metav1.DeleteOptions{},
				&metav1.CreateOptions{},
				&metav1.UpdateOptions{},
				&metav1.PatchOptions{},
				&metav1.Status{},
			)
			return nil
		})

	s := runtime.NewScheme()
	sb.AddToScheme(s)

	cm := corev1.ConfigMap{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(got.Unstructured.Object, &cm)
	if err != nil {
		log.Error("unable to convert unstructured to configmap",
			slog.Any("reference", ref), slog.Any("err", err))
		return map[string]string{}
	}

	return cm.Data
}
