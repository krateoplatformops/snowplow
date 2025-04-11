package widgets

import (
	"context"
	"encoding/json"
	"fmt"

	templatesv1 "github.com/krateoplatformops/snowplow/apis/templates/v1"
	"github.com/krateoplatformops/snowplow/internal/objects"
	"github.com/krateoplatformops/snowplow/internal/resolvers/restactions"
	xenv "github.com/krateoplatformops/snowplow/plumbing/env"
	"github.com/krateoplatformops/snowplow/plumbing/maps"
)

const (
	apiRefKey = "apiRef"
)

func resolveRESTActionRef(ctx context.Context, opts ResolveOptions) (map[string]any, error) {
	ref, err := getRESTActionRef(opts.In.Object)
	if err != nil {
		return nil, err
	}
	if ref.Name == "" || ref.Namespace == "" {
		return map[string]any{}, nil
	}

	res := objects.Get(ctx, ref)
	if res.Err != nil {
		return map[string]any{}, err
	}

	ra, err := convertToRESTAction(res.Unstructured.Object)
	if res.Err != nil {
		return map[string]any{}, err
	}

	raopts := restactions.ResolveOptions{
		In:         &ra,
		AuthnNS:    opts.AuthnNS,
		Username:   opts.Username,
		UserGroups: opts.UserGroups,
	}
	if xenv.TestMode() {
		raopts.SArc = opts.RC
	}

	if _, err = restactions.Resolve(ctx, raopts); err != nil {
		return map[string]any{}, err
	}

	return rawExtensionToMap(ra.Status)
}

func getRESTActionRef(in map[string]any) (objects.Reference, error) {
	api, ok, err := maps.NestedMapNoCopy(in, "spec", apiRefKey)
	if err != nil {
		return objects.Reference{}, err
	}
	if !ok {
		return objects.Reference{}, nil
	}

	dat, err := json.Marshal(api)
	if err != nil {
		return objects.Reference{}, err
	}

	ref := objects.Reference{
		Resource:   "restactions",
		APIVersion: fmt.Sprintf("%s/%s", templatesv1.Group, templatesv1.Version),
	}
	err = json.Unmarshal(dat, &ref)

	return ref, err
}

func convertToRESTAction(api map[string]any) (templatesv1.RESTAction, error) {
	dat, err := json.Marshal(api)
	if err != nil {
		return templatesv1.RESTAction{}, err
	}

	var ra templatesv1.RESTAction
	err = json.Unmarshal(dat, &ra)

	return ra, err
}
