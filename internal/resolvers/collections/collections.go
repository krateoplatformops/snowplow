package collections

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"

	"github.com/itchyny/gojq"
	"github.com/krateoplatformops/snowplow/apis"
	"github.com/krateoplatformops/snowplow/apis/templates/v1alpha1"
	templates "github.com/krateoplatformops/snowplow/apis/templates/v1alpha1"

	"github.com/krateoplatformops/snowplow/internal/objects"
	"github.com/krateoplatformops/snowplow/internal/resolvers/api"
	"github.com/krateoplatformops/snowplow/internal/resolvers/customforms"
	"github.com/krateoplatformops/snowplow/internal/resolvers/templaterefs"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/kubeutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"

	serializer "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

const (
	annotationKeyLastAppliedConfiguration = "kubectl.kubernetes.io/last-applied-configuration"
)

type ResolveOptions struct {
	In         *templates.Collection
	SArc       *rest.Config
	AuthnNS    string
	Username   string
	UserGroups []string
}

func Resolve(ctx context.Context, opts ResolveOptions) (*templates.Collection, error) {
	r := &collectionResolver{
		sarc:       opts.SArc,
		authnNS:    opts.AuthnNS,
		userName:   opts.Username,
		userGroups: opts.UserGroups,
	}
	return r.resolve(ctx, opts.In)
}

type collectionResolver struct {
	sarc       *rest.Config
	authnNS    string
	userName   string
	userGroups []string
}

func (r *collectionResolver) resolve(ctx context.Context, in *templates.Collection) (*templates.Collection, error) {
	if r.sarc == nil {
		var err error
		r.sarc, err = rest.InClusterConfig()
		if err != nil {
			return in, err
		}
	}

	log := xcontext.Logger(ctx)

	// Resolve API calls
	dict, err := api.Resolve(ctx, in.Spec.API, api.ResolveOptions{
		SARc:       r.sarc,
		AuthnNS:    r.authnNS,
		Username:   r.userName,
		UserGroups: r.userGroups,
	})
	if err != nil {
		return in, err
	}
	if dict == nil {
		dict = map[string]any{}
	}

	in.Status.Props = map[string]string{}
	if ref := in.Spec.PropsRef; ref != nil {
		var err error
		in.Status.Props, err = kubeutil.ConfigMapData(ctx, r.sarc, ref.Name, ref.Namespace)
		if err != nil {
			log.Error("unable resolve customform props",
				slog.String("name", ref.Name),
				slog.String("namespace", ref.Namespace),
				slog.Any("err", err))
			return in, err
		}
	}

	in.Status.UID = string(in.UID)
	in.Status.Name = in.Name
	in.Status.Type = in.Spec.Type

	in.Status.Items = []*runtime.RawExtension{}

	all := templaterefs.Expand(ctx, templaterefs.ExpandOptions{
		TemplateIterators: in.Spec.TemplateIterators,
		Dict:              dict,
	})

	scheme := runtime.NewScheme()
	if err := apis.AddToScheme(scheme); err != nil {
		return in, err
	}

	for _, el := range all {
		in.Status.Items = append(
			in.Status.Items,
			r.resolveReference(ctx, el, scheme),
		)
	}

	if in.Annotations != nil {
		delete(in.Annotations, annotationKeyLastAppliedConfiguration)
	}

	if in.ManagedFields != nil {
		in.ManagedFields = nil
	}

	return in, nil
}

func (r *collectionResolver) resolveReference(ctx context.Context, in *templates.ObjectReference, rs *runtime.Scheme) *runtime.RawExtension {
	if in == nil {
		return &runtime.RawExtension{Raw: []byte{}}
	}

	log := xcontext.Logger(ctx)

	gv, err := schema.ParseGroupVersion(in.APIVersion)
	if err != nil {
		log.Error("unable to parse group version", slog.Any("reference", in), slog.Any("err", err))
		return &runtime.RawExtension{Raw: []byte{}}
	}
	gvr := gv.WithResource(in.Resource)

	got := objects.Get(ctx, objects.Reference{
		Name: in.Name, Namespace: in.Namespace,
		APIVersion: gvr.GroupVersion().String(),
		Resource:   gvr.Resource,
	})
	if got.Err != nil {
		log.Error(got.Err.Message, slog.Any("reference", in), slog.Any("err", err))
		return &runtime.RawExtension{Raw: []byte{}}
	}

	var obj runtime.Object
	switch apis.GetTemplateKind(gvr.GroupResource()) {
	case apis.CustomFormTemplate:
		var cr v1alpha1.CustomForm
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(got.Unstructured.Object, &cr)
		if err != nil {
			log.Error("unable to convert custom form template from unstructured",
				slog.Any("reference", in), slog.Any("err", err))
			return &runtime.RawExtension{Raw: []byte{}}
		}

		ctx = xcontext.BuildContext(ctx, xcontext.WithJQTemplate())
		obj, err = customforms.Resolve(ctx, customforms.ResolveOptions{
			In:         &cr,
			Username:   r.userName,
			UserGroups: r.userGroups,
			AuthnNS:    r.authnNS,
		})
		if err != nil {
			log.Error("unable to resolve template reference", slog.Any("err", err))
			return &runtime.RawExtension{Raw: []byte{}}
		}
	default:
		log.Error("template reference resolution not implemented", slog.Any("reference", in))
		return &runtime.RawExtension{Raw: []byte{}}
	}

	buf := bytes.Buffer{}
	s := serializer.NewSerializerWithOptions(serializer.DefaultMetaFactory, rs, rs,
		serializer.SerializerOptions{Yaml: false, Pretty: true, Strict: false})
	err = s.Encode(obj, &buf)
	if err != nil {
		log.Error("unable to serialize object", slog.Any("reference", in), slog.Any("err", err))
		return &runtime.RawExtension{Raw: []byte{}}
	}

	dat, err := jq(".status | {status: .}", buf.Bytes())
	if err != nil {
		log.Error("unable to jq object status", slog.Any("reference", in), slog.Any("err", err))
		return &runtime.RawExtension{
			Raw: buf.Bytes(),
		}
	}

	return &runtime.RawExtension{
		Raw: dat,
	}
}

func jq(q string, buf []byte) ([]byte, error) {
	data := map[string]any{}
	err := json.NewDecoder(bytes.NewReader(buf)).Decode(&data)
	if err != nil {
		return nil, err
	}

	enc := newEncoder(false, 0)

	query, err := gojq.Parse(q)
	if err != nil {
		return nil, err
	}

	iter := query.Run(data) // or query.RunWithContext
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, err
		}
		if err := enc.encode(v); err != nil {
			return nil, err
		}
	}

	return enc.w.Bytes(), nil
}
