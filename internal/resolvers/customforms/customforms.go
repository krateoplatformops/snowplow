package customforms

import (
	"context"
	"fmt"
	"log/slog"

	templates "github.com/krateoplatformops/snowplow/apis/templates/v1"
	"github.com/krateoplatformops/snowplow/internal/resolvers/actions"
	"github.com/krateoplatformops/snowplow/internal/resolvers/api"
	app "github.com/krateoplatformops/snowplow/internal/resolvers/app/customforms"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/kubeutil"
	"github.com/krateoplatformops/snowplow/plumbing/ptr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

const (
	annotationKeyLastAppliedConfiguration = "kubectl.kubernetes.io/last-applied-configuration"
)

type ResolveOptions struct {
	In         *templates.CustomForm
	SArc       *rest.Config
	AuthnNS    string
	Username   string
	UserGroups []string
}

func Resolve(ctx context.Context, opts ResolveOptions) (*templates.CustomForm, error) {
	if opts.SArc == nil {
		var err error
		opts.SArc, err = rest.InClusterConfig()
		if err != nil {
			return opts.In, err
		}
	}

	log := xcontext.Logger(ctx)

	tpl := xcontext.JQTemplate(ctx)
	if tpl == nil {
		return opts.In, fmt.Errorf("missing jq template engine")
	}

	// Resolve 'in.Spec.PropsRef'
	opts.In.Status.Props = map[string]string{}
	if ref := opts.In.Spec.PropsRef; ref != nil {
		var err error
		opts.In.Status.Props, err = kubeutil.ConfigMapData(ctx, opts.SArc, ref.Name, ref.Namespace)
		if err != nil {
			log.Error("unable resolve customform props",
				slog.String("name", ref.Name),
				slog.String("namespace", ref.Namespace),
				slog.Any("err", err))
			return opts.In, err
		}
	}

	opts.In.Status.UID = ptr.To(string(opts.In.UID))
	opts.In.Status.Name = opts.In.Name
	opts.In.Status.Type = opts.In.Spec.Type

	// Resolve API calls
	dict, err := api.Resolve(ctx, opts.In.Spec.API, api.ResolveOptions{
		RC:         opts.SArc,
		AuthnNS:    opts.AuthnNS,
		Username:   opts.Username,
		UserGroups: opts.UserGroups,
	})
	if err != nil {
		return opts.In, err
	}
	if dict == nil {
		dict = map[string]any{}
	}

	// Resolve app
	schema, err := app.Resolve(ctx, opts.In.Spec.App.Template, dict)
	if err != nil {
		return opts.In, err
	}

	opts.In.Status.Content = &templates.CustomFormStatusContent{
		Schema: &runtime.RawExtension{
			Object: schema,
		},
	}

	// Resolve actions (eventually)
	if len(opts.In.Spec.Actions) == 0 {
		return opts.In, nil
	}

	exp := actions.Expand(ctx, map[string]any{}, opts.In.Spec.Actions...)
	all, err := actions.Resolve(ctx, exp)
	if err != nil {
		return opts.In, err
	}

	opts.In.Status.Actions = make([]*templates.ActionResultTemplate, 0, len(all))
	for _, el := range all {
		if el.Payload != nil && el.Payload.MetaData != nil {
			el.Payload.MetaData.Name = opts.In.Spec.Type
		}

		opts.In.Status.Actions = append(opts.In.Status.Actions,
			&templates.ActionResultTemplate{
				Template: el,
			})
	}

	if opts.In.Annotations != nil {
		delete(opts.In.Annotations, annotationKeyLastAppliedConfiguration)
	}
	if opts.In.ManagedFields != nil {
		opts.In.ManagedFields = nil
	}

	return opts.In, nil
}
