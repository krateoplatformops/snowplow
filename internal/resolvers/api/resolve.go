package api

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	xcontext "github.com/krateoplatformops/plumbing/context"
	httpcall "github.com/krateoplatformops/plumbing/http/request"
	"github.com/krateoplatformops/plumbing/http/response"
	templates "github.com/krateoplatformops/snowplow/apis/templates/v1"
	"k8s.io/client-go/rest"
)

const (
	annotationKeyVerboseAPI = "krateo.io/verbose"
	headerAcceptJSON        = "Accept: application/json"
)

type ResolveOptions struct {
	RC         *rest.Config
	AuthnNS    string
	Username   string
	UserGroups []string
	Verbose    bool
	Items      []*templates.API
}

func Resolve(ctx context.Context, opts ResolveOptions) map[string]any {
	if len(opts.Items) == 0 {
		return map[string]any{}
	}

	if opts.RC == nil {
		var err error
		opts.RC, err = rest.InClusterConfig()
		if err != nil {
			return map[string]any{}
		}
	}

	log := xcontext.Logger(ctx)

	// Sort API by Depends
	names, err := topologicalSort(opts.Items)
	if err != nil {
		log.Error("unable to sorted api by deps", slog.Any("error", err))
		return map[string]any{}
	}
	log.Debug("sorted api by deps", slog.Any("names", names))

	apiMap := make(map[string]*templates.API, len(opts.Items))
	for _, id := range names {
		for _, el := range opts.Items {
			if el.Name == id {
				apiMap[id] = el
				break
			}
		}
	}
	log.Debug("created api map", slog.Int("total", len(apiMap)))

	// Endpoints reference mapper
	mapper := endpointReferenceMapper{
		authnNS:  opts.AuthnNS,
		username: opts.Username,
		rc:       opts.RC,
	}

	dict := map[string]any{}

	for _, id := range names {
		// Get the api with this identifier
		apiCall, ok := apiMap[id]
		if !ok {
			log.Warn("api not found in apiMap", slog.Any("name", id))
			continue
		}

		// Add Krateo HTTP Request headers
		if apiCall.Headers == nil {
			apiCall.Headers = []string{headerAcceptJSON}
		}
		apiCall.Headers = append(apiCall.Headers,
			fmt.Sprintf("X-Krateo-User: %s", opts.Username))
		apiCall.Headers = append(apiCall.Headers,
			fmt.Sprintf("X-Krateo-Groups: %s", strings.Join(opts.UserGroups, ",")))

		// Resolve the endpoint
		ep, err := mapper.resolveOne(ctx, apiCall.EndpointRef)
		if err != nil {
			log.Error("unable to resolve api endpoint reference",
				slog.String("name", id), slog.Any("error", err))
			return dict
		}
		if opts.Verbose {
			ep.Debug = opts.Verbose
		}
		log.Debug("resolved endpoint for api call",
			slog.String("name", id), slog.String("host", ep.ServerURL))

		tmp := createRequestOptions(ctx, apiCall, dict)
		if len(tmp) == 0 {
			log.Warn("empty request options for http call", slog.Any("name", id))
			continue
		}

		for _, call := range tmp {
			call.Endpoint = &ep
			call.ResponseHandler = jsonHandler(ctx, jsonHandlerOptions{
				key: id, out: dict, filter: apiCall.Filter,
			})

			log.Debug("calling api", slog.String("name", id),
				slog.String("host", call.Endpoint.ServerURL), slog.String("path", call.Path))

			res := httpcall.Do(ctx, call)
			if res.Status == response.StatusFailure {
				log.Error("api call response failure", slog.String("name", id),
					slog.String("host", call.Endpoint.ServerURL), slog.String("path", call.Path),
					slog.String("error", res.Message))

				tmp, err := response.AsMap(res)
				if err != nil {
					log.Warn("unable to encode status as dict", slog.Any("err", err))
				}

				if len(tmp) > 0 {
					dict[call.ErrorKey] = tmp
				} else {
					dict[call.ErrorKey] = res.Message
				}

				if call.ContinueOnError {
					continue
				}

				return dict
			}
		}
	}

	return dict
}
