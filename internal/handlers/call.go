package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	xcontext "github.com/krateoplatformops/plumbing/context"
	"github.com/krateoplatformops/plumbing/http/request"
	"github.com/krateoplatformops/plumbing/http/response"
	"github.com/krateoplatformops/plumbing/ptr"
	"github.com/krateoplatformops/snowplow/internal/handlers/util"
	"github.com/krateoplatformops/snowplow/internal/telemetry"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func Call(verbose bool, metrics *telemetry.Metrics) http.Handler {
	return &callHandler{
		verbose: verbose,
		metrics: metrics,
	}
}

var _ http.Handler = (*callHandler)(nil)

type callHandler struct {
	verbose bool
	metrics *telemetry.Metrics
}

// @Summary Call Endpoint
// @Description Handle Resources
// @ID call
// @Param  apiVersion       query   string  true  "Resource API Group and Version"
// @Param  resource         query   string  true  "Resource Plural"
// @Param  name             query   string  true  "Resource name"
// @Param  namespace        query   string  true  "Resource namespace"
// @Param  page             query   string  false "Pagination desired page"
// @Param  perPage          query   string  false "Pagination desired per page items"
// @Param  extras           query   string  false "JSON encoded map of extra params"
// @Param data body string false "Object"
// @Produce  json
// @Success 200 {object} map[string]any
// @Failure 400 {object} response.Status
// @Failure 401 {object} response.Status
// @Failure 404 {object} response.Status
// @Failure 500 {object} response.Status
// @Router /call [get]
// @Router /call [post]
// @Router /call [put]
// @Router /call [patch]
// @Router /call [delete]
func (r *callHandler) ServeHTTP(wri http.ResponseWriter, req *http.Request) {
	totalStart := time.Now()
	method := strings.ToUpper(req.Method)
	apiGroup := ""
	resource := ""

	validateStart := time.Now()
	opts, err := r.validateRequest(req)
	r.metrics.RecordCallStageDuration(req.Context(), "validate", method, apiGroup, resource, time.Since(validateStart))
	if err != nil {
		r.metrics.IncCallError(req.Context(), "validate", method, apiGroup, resource, http.StatusBadRequest)
		r.metrics.RecordCallRequest(req.Context(), method, apiGroup, resource, http.StatusBadRequest, time.Since(totalStart))
		response.BadRequest(wri, err)
		return
	}
	apiGroup = opts.gvr.Group
	resource = opts.gvr.Resource

	buildURIStart := time.Now()
	uri, err := buildURIPath(opts)
	r.metrics.RecordCallStageDuration(req.Context(), "build_uri", method, apiGroup, resource, time.Since(buildURIStart))
	if err != nil {
		r.metrics.IncCallError(req.Context(), "build_uri", method, apiGroup, resource, http.StatusInternalServerError)
		r.metrics.RecordCallRequest(req.Context(), method, apiGroup, resource, http.StatusInternalServerError, time.Since(totalStart))
		response.InternalError(wri, err)
		return
	}

	log := xcontext.Logger(req.Context())

	start := time.Now()

	userConfigStart := time.Now()
	ep, err := xcontext.UserConfig(req.Context())
	r.metrics.RecordCallStageDuration(req.Context(), "user_config", method, apiGroup, resource, time.Since(userConfigStart))
	if err != nil {
		log.Error("unable to get user endpoint", slog.Any("err", err))
		r.metrics.IncCallError(req.Context(), "user_config", method, apiGroup, resource, http.StatusUnauthorized)
		r.metrics.RecordCallRequest(req.Context(), method, apiGroup, resource, http.StatusUnauthorized, time.Since(totalStart))
		response.Unauthorized(wri, err)
		return
	}
	ep.Debug = r.verbose

	log.Debug("user config succesfully loaded", slog.Any("endpoint", ep))

	dict := map[string]any{}
	callOpts := request.RequestOptions{
		RequestInfo: request.RequestInfo{
			Path: uri,
			Verb: ptr.To(strings.ToUpper(opts.verb)),
			Headers: []string{
				"Accept: application/json",
			},
		},
		Endpoint:        &ep,
		ResponseHandler: callResponseHandler(dict),
	}
	if opts.dat != nil && has([]string{http.MethodPost, http.MethodPut, http.MethodPatch}, opts.verb) {
		callOpts.Headers = append(callOpts.Headers,
			fmt.Sprintf("Content-Type: %s", opts.contentType),
		)
		callOpts.Payload = ptr.To(string(opts.dat))
	}

	upstreamStart := time.Now()
	rt := request.Do(req.Context(), callOpts)
	r.metrics.RecordCallStageDuration(req.Context(), "upstream", method, apiGroup, resource, time.Since(upstreamStart))
	if rt.Status == response.StatusFailure {
		log.Error("unable to call endpoint",
			slog.String("verb", strings.ToUpper(opts.verb)),
			slog.String("uri", uri),
			slog.String("err", rt.Message))
		r.metrics.IncCallError(req.Context(), "upstream", method, apiGroup, resource, rt.Code)
		r.metrics.RecordCallRequest(req.Context(), method, apiGroup, resource, rt.Code, time.Since(totalStart))
		response.Encode(wri, rt)
		return
	}

	log.Info("endpoint call done",
		slog.String("verb", strings.ToUpper(opts.verb)),
		slog.String("uri", uri),
		slog.String("duration", util.ETA(start)),
	)

	wri.Header().Set("Content-Type", "application/json")
	wri.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(wri)
	enc.SetIndent("", "  ")
	if err := enc.Encode(dict); err != nil {
		log.Error("unable to serve api call response", slog.Any("err", err))
		r.metrics.IncCallError(req.Context(), "encode_response", method, apiGroup, resource, http.StatusInternalServerError)
		return
	}

	r.metrics.RecordCallRequest(req.Context(), method, apiGroup, resource, http.StatusOK, time.Since(totalStart))
}

func (r *callHandler) validateRequest(req *http.Request) (opts callOptions, err error) {
	opts.verb = req.Method
	if has([]string{http.MethodPost, http.MethodPut, http.MethodPatch}, opts.verb) {
		opts.contentType = req.Header.Get("Content-type")
		if opts.contentType == "" {
			opts.contentType = "application/json"
		}
	}

	opts.gvr, err = util.ParseGVR(req)
	if err != nil {
		return
	}

	opts.nsn, err = util.ParseNamespacedName(req)
	if err != nil {
		return
	}

	if val := req.URL.Query().Get("perPage"); val != "" {
		opts.perPage, err = strconv.Atoi(val)
		if err != nil {
			return
		}
	}

	if val := req.URL.Query().Get("page"); val != "" {
		opts.page, err = strconv.Atoi(val)
		if err != nil {
			return
		}
	}

	if val := req.URL.Query().Get("cursor"); val != "" {
		opts.page = -1
		opts.cursor = val
	}

	if req.Body != nil {
		opts.dat, err = io.ReadAll(io.LimitReader(req.Body, 1048576))
		if err != nil {
			return
		}
	}

	return
}

type callOptions struct {
	gvr         schema.GroupVersionResource
	nsn         types.NamespacedName
	verb        string
	contentType string
	perPage     int
	page        int
	cursor      string
	dat         []byte
}

func buildURIPath(opts callOptions) (string, error) {
	base := path.Join("/apis", opts.gvr.Group, opts.gvr.Version)
	if len(opts.gvr.Group) == 0 {
		base = path.Join("/api", opts.gvr.Version)
	}

	uriPath := path.Join(base, "namespaces", opts.nsn.Namespace, opts.gvr.Resource)
	if strings.EqualFold("namespaces", opts.gvr.Resource) {
		uriPath = path.Join(base, opts.gvr.Resource)
	}

	if has([]string{
		http.MethodDelete,
		http.MethodGet,
		http.MethodPut,
		http.MethodPatch,
	}, opts.verb) {
		uriPath = path.Join(uriPath, opts.nsn.Name)
	}

	// Aggiunta dei query parametri, se necessario
	query := url.Values{}
	if opts.perPage > 0 {
		query.Set("perPage", strconv.Itoa(opts.perPage))
	}
	if opts.page > 0 {
		query.Set("page", strconv.Itoa(opts.page))
	}
	if opts.cursor != "" {
		query.Set("cursor", opts.cursor)
	}

	if len(query) > 0 {
		uriPath += "?" + query.Encode()
	}

	return uriPath, nil
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
