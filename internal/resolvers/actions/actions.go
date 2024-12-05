package actions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	templates "github.com/krateoplatformops/snowplow/apis/templates/v1alpha1"
	"github.com/krateoplatformops/snowplow/internal/dynamic"
	"github.com/krateoplatformops/snowplow/internal/rbac"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/kubeconfig"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

func Resolve(ctx context.Context, actions []*templates.Action) (all []*templates.ActionResult, err error) {
	log := xcontext.Logger(ctx)

	ep, err := xcontext.UserConfig(ctx)
	if err != nil {
		log.Error("unable to get user endpoint", slog.Any("err", err))
		return nil, err
	}

	rc, err := kubeconfig.NewClientConfig(ctx, ep)
	if err != nil {
		log.Error("unable to create user client config", slog.Any("err", err))
		return nil, err
	}

	for _, el := range actions {
		if el.Template == nil {
			continue
		}

		res, err2 := resolveOne(ctx, rc, el.Template)
		if err2 != nil {
			err = errors.Join(err, err2)
			continue
		}

		all = append(all, res...)
	}

	return
}

func resolveOne(ctx context.Context, rc *rest.Config, in *templates.ActionTemplate) ([]*templates.ActionResult, error) {
	all := []*templates.ActionResult{}
	if in == nil {
		return all, nil
	}

	gv, err := schema.ParseGroupVersion(in.APIVersion)
	if err != nil {
		return all, err
	}
	gvr := gv.WithResource(in.Resource)

	gvk, err := dynamic.KindFor(rc, gvr)
	if err != nil {
		return all, err
	}

	verbs := mapVerbs(in)
	for _, verb := range verbs {
		ok := rbac.UserCan(ctx, rbac.UserCanOptions{
			Verb:          verb,
			GroupResource: gvr.GroupResource(),
			Namespace:     in.Namespace,
		})
		if !ok {
			xcontext.Logger(ctx).Error("action not allowed",
				slog.String("kind", gvk.Kind),
				slog.String("resource", gvr.Resource),
				slog.String("namespace", in.Namespace))
			continue
		}

		el := &templates.ActionResult{
			ID:   in.ID,
			Verb: kubeToREST[verb],
			Path: fmt.Sprintf("/call?resource=%s&apiVersion=%s&name=%s&namespace=%s",
				gvr.Resource, gvr.GroupVersion().String(), in.Name, in.Namespace),
		}

		if tot := len(in.PayloadToOverride); tot > 0 {
			el.PayloadToOverride = make([]templates.Data, tot)
			copy(el.PayloadToOverride, in.PayloadToOverride)
		}

		el.Payload = &templates.ActionResultPayload{
			Kind:       gvk.Kind,
			APIVersion: in.APIVersion,
			MetaData: &templates.Reference{
				Name:      in.Name,
				Namespace: in.Namespace,
			},
		}

		all = append(all, el)
	}

	return all, nil
}
