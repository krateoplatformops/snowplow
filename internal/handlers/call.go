package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path"
	"strings"

	"github.com/krateoplatformops/snowplow/internal/handlers/util"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/env"
	"github.com/krateoplatformops/snowplow/plumbing/http/request"
	"github.com/krateoplatformops/snowplow/plumbing/http/response"
	"github.com/krateoplatformops/snowplow/plumbing/ptr"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func Call() http.Handler {
	return &callHandler{
		authnNS: env.String("AUTHN_NAMESPACE", ""),
		verbose: env.True("DEBUG"),
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
		response.BadRequest(wri, err)
		return
	}

	uri, err := buildURI(opts)
	if err != nil {
		response.InternalError(wri, err)
		return
	}

	log := xcontext.Logger(req.Context())

	ep, err := xcontext.UserConfig(req.Context())
	if err != nil {
		log.Error("unable to get user endpoint", slog.Any("err", err))
		response.Unauthorized(wri, err)
		return
	}
	ep.Debug = r.verbose

	log.Debug("user config succesfully loaded", slog.Any("endpoint", ep))

	dict := map[string]any{}
	callOpts := request.RequestOptions{
		Path: ptr.To(uri),
		Verb: ptr.To(strings.ToUpper(opts.verb)),
		Headers: []string{
			"Accept: application/json",
		},
		Endpoint:        &ep,
		ResponseHandler: callResponseHandler(dict),
	}
	if has([]string{http.MethodPost, http.MethodPut}, opts.verb) {
		callOpts.Payload = ptr.To(string(opts.dat))
	}

	rt := request.Do(req.Context(), callOpts)
	if rt.Status == response.StatusFailure {
		log.Error("unable to call endpoint",
			slog.String("verb", strings.ToUpper(opts.verb)),
			slog.String("uri", uri),
			slog.String("err", rt.Message))
		response.Encode(wri, rt)
		return
	}

	wri.Header().Set("Content-Type", "application/json")
	wri.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(wri)
	enc.SetIndent("", "  ")
	if err := enc.Encode(dict); err != nil {
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

	opts.gvr, err = util.ParseGVR(req)
	if err != nil {
		return
	}

	opts.nsn, err = util.ParseNamespacedName(req)
	if err != nil {
		return
	}

	opts.dat, err = io.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		return
	}

	return
}

type callOptions struct {
	gvr     schema.GroupVersionResource
	nsn     types.NamespacedName
	verb    string
	subject string
	groups  []string
	dat     []byte
}

func buildURI(opts callOptions) (string, error) {
	base := path.Join("/apis", opts.gvr.Group, opts.gvr.Version)
	if len(opts.gvr.Group) == 0 {
		base = path.Join("/api", opts.gvr.Version)
	}

	uri := path.Join(base, "namespaces", opts.nsn.Namespace, opts.gvr.Resource)
	if strings.EqualFold("namespaces", opts.gvr.Resource) {
		uri = path.Join(base, opts.gvr.Resource)
	}

	if has([]string{http.MethodDelete, http.MethodGet, http.MethodPut}, opts.verb) {
		uri = path.Join(uri, opts.nsn.Name)
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

func callResponseHandler(out map[string]any) func(io.ReadCloser) error {
	return func(in io.ReadCloser) error {
		dat, err := io.ReadAll(in)
		if err != nil {
			return err
		}

		x := bytes.TrimSpace(dat)
		isArray := len(x) > 0 && x[0] == '['

		if isArray {
			v := []any{}
			err := json.Unmarshal(dat, &v)
			if err != nil {
				return err
			}
			out["items"] = v
			return nil
		}

		return json.Unmarshal(dat, &out)
	}
}
