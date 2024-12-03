package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path"
	"strings"

	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/env"
	"github.com/krateoplatformops/snowplow/plumbing/http/request"
	"github.com/krateoplatformops/snowplow/plumbing/http/response/status"
	"github.com/krateoplatformops/snowplow/plumbing/ptr"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func Call() http.Handler {
	return &callHandler{
		authnNS: env.String(env.AuthnNamespace, "krateo-system"),
		verbose: env.Bool("DEBUG", false),
	}
}

var _ http.Handler = (*callHandler)(nil)

type callHandler struct {
	authnNS string
	verbose bool
}

// @Summary Call Endpoint
// @Description Handle Resources
// @ID call
// @Param  X-Krateo-User    header  string  true  "Krateo User"
// @Param  X-Krateo-Groups  header  string  true  "Krateo User Groups"
// @Param  apiVersion       query   string  true  "Resource API Group and Version"
// @Param  resource         query   string  true  "Resource Plural"
// @Param  name             query   string  true  "Resource name"
// @Param  namespace        query   string  true  "Resource namespace"
// @Param data body string false "Object"
// @Produce  json
// @Success 200 {object} map[string]any
// @Failure 400 {object} status.Status
// @Failure 401 {object} status.Status
// @Failure 404 {object} status.Status
// @Failure 500 {object} status.Status
// @Router /call [get]
// @Router /call [post]
// @Router /call [put]
// @Router /call [delete]
func (r *callHandler) ServeHTTP(wri http.ResponseWriter, req *http.Request) {
	opts, err := r.validateRequest(req)
	if err != nil {
		status.BadRequest(wri, err)
		return
	}

	uri, err := buildURI(opts)
	if err != nil {
		status.InternalError(wri, err)
		return
	}

	log := xcontext.Logger(req.Context())

	ep, err := xcontext.UserConfig(req.Context())
	if err != nil {
		log.Error("unable to get user endpoint", slog.Any("err", err))
		status.Unauthorized(wri, err)
		return
	}
	ep.Debug = env.Bool("DEBUG", false)
	ep.ServerURL = "https://kubernetes.default.svc"

	callOpts := request.Options{
		Path: ptr.To(uri),
		Verb: ptr.To(strings.ToUpper(opts.verb)),
		Headers: []string{
			"Accept: application/json",
		},
		Endpoint: &ep,
	}
	if has([]string{http.MethodPost, http.MethodPut}, opts.verb) {
		callOpts.Payload = ptr.To(string(opts.dat))
	}

	rt := request.Do(req.Context(), callOpts)
	if rt.Status.Status == status.StatusFailure {
		log.Error("unable to call endpoint",
			slog.String("verb", strings.ToUpper(opts.verb)),
			slog.String("uri", uri),
			slog.String("err", rt.Status.Message))
		status.Encode(wri, rt.Status)
		return
	}

	wri.Header().Set("Content-Type", "application/json")
	wri.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(wri)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rt.Map); err != nil {
		log.Error("unable to serve api call response", slog.Any("err", err))
	}
}

func (r *callHandler) validateRequest(req *http.Request) (opts callOptions, err error) {
	opts.verb = req.Method

	opts.subject = req.Header.Get(xcontext.LabelKrateoUser)
	if len(opts.subject) == 0 {
		err = fmt.Errorf("missing '%s' header", xcontext.LabelKrateoUser)
		return
	}

	opts.groups = strings.Split(req.Header.Get(xcontext.LabelKrateoGroups), ",")
	if len(opts.groups) == 0 {
		err = fmt.Errorf("missing '%s' header", xcontext.LabelKrateoGroups)
		return
	}

	opts.apiVersion = req.URL.Query().Get("apiVersion")
	if len(opts.apiVersion) == 0 {
		err = fmt.Errorf("missing 'apiVersion' query parameter")
		return
	}

	opts.resource = req.URL.Query().Get("resource")
	if len(opts.resource) == 0 {
		err = fmt.Errorf("missing 'resource' query parameter")
		return
	}

	opts.name = req.URL.Query().Get("name")
	if len(opts.name) == 0 {
		err = fmt.Errorf("missing 'name' query parameter")
		return
	}

	opts.namespace = req.URL.Query().Get("namespace")
	if len(opts.namespace) == 0 {
		err = fmt.Errorf("missing 'namespace' query parameter")
		return
	}

	opts.dat, err = io.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		return
	}

	return
}

type callOptions struct {
	apiVersion string
	resource   string
	name       string
	namespace  string
	verb       string
	subject    string
	groups     []string
	dat        []byte
}

func buildURI(opts callOptions) (string, error) {
	gv, err := schema.ParseGroupVersion(opts.apiVersion)
	if err != nil {
		return "", err
	}

	base := path.Join("/apis", gv.Group, gv.Version)
	if len(gv.Group) == 0 {
		base = path.Join("/api", gv.Version)
	}

	uri := path.Join(base, "namespaces", opts.namespace, opts.resource)
	if strings.EqualFold("namespaces", opts.resource) {
		uri = path.Join(base, opts.resource)
	}

	if has([]string{http.MethodDelete, http.MethodGet, http.MethodPut}, opts.verb) {
		uri = path.Join(uri, opts.name)
	}

	return uri, nil
}

func has(s []string, e string) bool {
	for _, a := range s {
		if strings.EqualFold(a, e) {
			return true
		}
	}

	return false
}
