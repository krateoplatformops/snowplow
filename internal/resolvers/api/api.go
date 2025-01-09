package api

import (
	"context"
	"fmt"
	"strings"

	templates "github.com/krateoplatformops/snowplow/apis/templates/v1alpha1"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/deps"
	"github.com/krateoplatformops/snowplow/plumbing/endpoints"
	"github.com/krateoplatformops/snowplow/plumbing/env"
	httpcall "github.com/krateoplatformops/snowplow/plumbing/http/request"
	httpstatus "github.com/krateoplatformops/snowplow/plumbing/http/response/status"
	"github.com/krateoplatformops/snowplow/plumbing/kubeutil"
	"github.com/krateoplatformops/snowplow/plumbing/ptr"
	"k8s.io/client-go/rest"
)

type ResolveOptions struct {
	SARc       *rest.Config
	AuthnNS    string
	Username   string
	UserGroups []string
}

func Resolve(ctx context.Context, apiList []*templates.API, opts ResolveOptions) (dict map[string]any, err error) {
	if len(apiList) == 0 {
		return
	}

	tpl := xcontext.JQTemplate(ctx)
	if tpl == nil {
		return nil, fmt.Errorf("missing jq template engine")
	}

	// Resolve all endpoints references
	endpointMap := map[string]endpoints.Endpoint{}
	for _, el := range apiList {
		isInternal := false
		if el.EndpointRef == nil {
			el.EndpointRef = &templates.Reference{
				Namespace: opts.AuthnNS,
				Name:      fmt.Sprintf("%s-clientconfig", kubeutil.MakeDNS1123Compatible(opts.Username)),
			}
			isInternal = true
		}

		ep, err := endpoints.FromSecret(ctx, opts.SARc, el.EndpointRef.Name, el.EndpointRef.Namespace)
		if err != nil {
			return nil, err
		}
		if isInternal && !env.TestMode() {
			ep.ServerURL = "https://kubernetes.default.svc"
		}

		endpointMap[el.Name] = ep
	}

	dict = map[string]any{}

	// Sort API by dependencies
	apiMap := sortApiByDeps(apiList)

	for name, api := range apiMap {
		ep, ok := endpointMap[name]
		if !ok {
			return dict, fmt.Errorf("endpoint for api %q not found; skipping api call", name)
		}
		ep.Debug = env.True("DEBUG")

		if pt := ptr.Deref(api.Path, ""); len(pt) > 0 && len(dict) > 0 {
			rt, err := tpl.Execute(pt, dict)
			if err != nil {
				return nil, err
			}
			api.Path = ptr.To(rt)
		}

		if api.Headers == nil {
			api.Headers = []string{}
		}
		api.Headers = append(api.Headers,
			fmt.Sprintf("X-Krateo-User: %s", opts.Username))
		api.Headers = append(api.Headers,
			fmt.Sprintf("X-Krateo-Groups: %s", strings.Join(opts.UserGroups, ",")))

		res := httpcall.Do(ctx, httpcall.Options{
			Path:     api.Path,
			Verb:     api.Verb,
			Headers:  api.Headers,
			Payload:  api.Payload,
			Endpoint: &ep,
		})
		if res.Status.Status == httpstatus.StatusFailure {
			return dict, fmt.Errorf("unable to perform api call %q: %s", api.Name, res.Status.Message)
		}

		dict[name] = res.Map
	}

	return
}

func sortApiByDeps(items []*templates.API) map[string]*templates.API {
	g := deps.New()

	nodep := []string{}
	for _, el := range items {
		dep := ptr.Deref(el.DependOn, "")
		if len(dep) == 0 {
			nodep = append(nodep, el.Name)
			continue
		}
		_ = g.DependOn(el.Name, dep)
	}

	all := append(nodep, g.TopoSorted()...)

	apiMap := make(map[string]*templates.API, len(all))
	for _, name := range all {
		for _, x := range items {
			if x.Name == name {
				apiMap[name] = x
				break
			}
		}
	}
	return apiMap
}
