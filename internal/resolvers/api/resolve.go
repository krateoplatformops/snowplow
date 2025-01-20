package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	templates "github.com/krateoplatformops/snowplow/apis/templates/v1"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	httpcall "github.com/krateoplatformops/snowplow/plumbing/http/request"
	"github.com/krateoplatformops/snowplow/plumbing/http/response"
	"github.com/krateoplatformops/snowplow/plumbing/ptr"
	"k8s.io/client-go/rest"
)

const (
	annotationKeyVerboseAPI = "krateo.io/verbose"
)

type ResolveOptions struct {
	RC         *rest.Config
	AuthnNS    string
	Username   string
	UserGroups []string
	Verbose    bool
	Items      []*templates.API
}

func Resolve(ctx context.Context, opts ResolveOptions) (dict map[string]any, err error) {
	if len(opts.Items) == 0 {
		return
	}

	if opts.RC == nil {
		var err error
		opts.RC, err = rest.InClusterConfig()
		if err != nil {
			return map[string]any{}, err
		}
	}

	tpl := xcontext.JQTemplate(ctx)
	if tpl == nil {
		return map[string]any{}, fmt.Errorf("missing jq template engine")
	}

	log := xcontext.Logger(ctx)

	// Sort API by Depends
	names, err := topologicalSort(opts.Items)
	if err != nil {
		return map[string]any{}, err
	}
	log.Debug("topological sorted api calls", slog.Any("names", names))

	apiMap := make(map[string]*templates.API, len(names))
	for _, name := range names {
		for _, el := range opts.Items {
			if el.Name == name {
				apiMap[name] = el
				break
			}
		}
	}

	mustExpand := func(x *templates.API) bool {
		it := ptr.Deref(x.Iterator, "")
		return len(it) > 0
	}

	dict = map[string]any{}

	for _, el := range apiMap {
		// Resolve Endpoint Reference
		ep, err := resolveEndpointReference(ctx, resolveEndpointReferenceOptions{
			RC: opts.RC, AuthnNS: opts.AuthnNS, Username: opts.Username, Reference: el.EndpointRef,
		})
		if err != nil {
			log.Error("unable to resolve endpoint reference",
				slog.String("api", el.Name),
				slog.Any("reference", el.EndpointRef))

			return dict, err
		}

		if opts.Verbose {
			ep.Debug = opts.Verbose
		}

		// Add Krateo HTTP Request headers
		if el.Headers == nil {
			el.Headers = []string{"Accept: application/json"}
		}
		el.Headers = append(el.Headers,
			fmt.Sprintf("X-Krateo-User: %s", opts.Username))
		el.Headers = append(el.Headers,
			fmt.Sprintf("X-Krateo-Groups: %s", strings.Join(opts.UserGroups, ",")))

		batch := []*templates.API{}
		if mustExpand(el) {
			// enc := json.NewEncoder(os.Stderr)
			// enc.SetIndent("", "  ")
			// enc.Encode(dict)

			tmp := expandIterator(ctx, el, dict)
			if len(tmp) > 0 {
				batch = append(batch, tmp...)
			}
		} else {
			batch = append(batch, el)
			// Eval API path JQ expression eventually
			if len(el.Path) > 0 && len(dict) > 0 {
				rt, err := tpl.Execute(el.Path, dict)
				if err != nil {
					return nil, err
				}
				el.Path = rt
			}
		}

		// Call the batch API
		for _, y := range batch {
			log.Debug("calling api", slog.String("name", y.Name), slog.String("path", y.Path))

			res := httpcall.Do(ctx, httpcall.RequestOptions{
				Endpoint:        &ep,
				Path:            ptr.To(y.Path),
				Verb:            y.Verb,
				Headers:         y.Headers,
				Payload:         y.Payload,
				ResponseHandler: jsonResponseHandler(ctx, y.Name, dict, y.Filter),
			})
			if res.Status == response.StatusFailure {
				return dict, fmt.Errorf("unable to perform api call %q: %s", y.Name, res.Message)
			}
		}
	}

	/*
		if len(opts.Items) > 1 {
			lastElement := opts.Items[len(opts.Items)-1]
			dict = filterMapByKey(dict, lastElement.Name)
		}
	*/

	fmt.Printf("\n\n\n")
	fout, err := os.Create("temp.json")
	if err == nil {
		defer fout.Close()
		enc := json.NewEncoder(fout)
		enc.SetIndent("", "  ")
		enc.Encode(dict)
	}

	fmt.Printf("\n\n\n")

	return
}
