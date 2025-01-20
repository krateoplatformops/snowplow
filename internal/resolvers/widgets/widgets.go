package widgets

import (
	"context"
	"fmt"
	"log/slog"

	templates "github.com/krateoplatformops/snowplow/apis/templates/v1"
	"github.com/krateoplatformops/snowplow/internal/resolvers/actions"
	"github.com/krateoplatformops/snowplow/internal/resolvers/api"
	app "github.com/krateoplatformops/snowplow/internal/resolvers/app/widgets"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/kubeutil"

	"k8s.io/client-go/rest"
)

const (
	annotationKeyLastAppliedConfiguration = "kubectl.kubernetes.io/last-applied-configuration"
)

type ResolveOptions struct {
	In         *templates.Widget
	SArc       *rest.Config
	AuthnNS    string
	Username   string
	UserGroups []string
}

func Resolve(ctx context.Context, opts ResolveOptions) (*templates.Widget, error) {
	if opts.SArc == nil {
		var err error
		opts.SArc, err = rest.InClusterConfig()
		if err != nil {
			log := xcontext.Logger(ctx)
			log.Error("unable to get serviceaccount RESTCconfig",
				slog.String("name", opts.In.Name),
				slog.String("namespace", opts.In.Namespace),
				slog.Any("err", err))
			return opts.In, err
		}
	}

	tpl := xcontext.JQTemplate(ctx)
	if tpl == nil {
		log := xcontext.Logger(ctx)
		log.Error("unable to find jq template engine in context",
			slog.String("name", opts.In.Name),
			slog.String("namespace", opts.In.Namespace))
		return opts.In, fmt.Errorf("unable to find jq template engine in context")
	}

	// Resolve 'in.Spec.PropsRef'
	opts.In.Status.Props = map[string]string{}
	if ref := opts.In.Spec.PropsRef; ref != nil {
		var err error
		opts.In.Status.Props, err = kubeutil.ConfigMapData(ctx, opts.SArc, ref.Name, ref.Namespace)
		if err != nil {
			log := xcontext.Logger(ctx)
			log.Error("unable resolve widget props",
				slog.String("name", opts.In.Name),
				slog.String("namespace", opts.In.Namespace),
				slog.Any("err", err))
			return opts.In, err
		}
	}

	opts.In.Status.UID = string(opts.In.UID)
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

	traits := app.Resolve(ctx, opts.In.Spec.App, dict)

	exp := actions.Expand(ctx, map[string]any{}, opts.In.Spec.Actions...)
	all, err := actions.Resolve(ctx, exp)
	if err != nil {
		log := xcontext.Logger(ctx)
		log.Error("unable resolve widget actions",
			slog.String("name", opts.In.Name),
			slog.String("namespace", opts.In.Namespace),
			slog.Any("err", err))
		return opts.In, err
	}

	opts.In.Status.Items = make([]*templates.WidgetResult, len(traits))
	for i, el := range traits {
		opts.In.Status.Items[i] = &templates.WidgetResult{
			Traits:  el,
			Actions: all,
		}
	}

	if opts.In.Annotations != nil {
		delete(opts.In.Annotations, annotationKeyLastAppliedConfiguration)
	}
	if opts.In.ManagedFields != nil {
		opts.In.ManagedFields = nil
	}

	return opts.In, nil
}
