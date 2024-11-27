package request

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/krateoplatformops/snowplow/plumbing/endpoints"
	"github.com/krateoplatformops/snowplow/plumbing/http/response/status"
	"github.com/krateoplatformops/snowplow/plumbing/ptr"
)

const maxUnstructuredResponseTextBytes = 2048

type Options struct {
	Path     *string
	Verb     *string
	Headers  []string
	Payload  *string
	Endpoint *endpoints.Endpoint
}

type Result struct {
	Map    map[string]any
	Status status.Status
}

func Do(ctx context.Context, opts Options) (res Result) {
	uri := strings.TrimSuffix(opts.Endpoint.ServerURL, "/")
	if pt := ptr.Deref(opts.Path, ""); len(pt) > 0 {
		uri = fmt.Sprintf("%s/%s", uri, strings.TrimPrefix(pt, "/"))
	}

	u, err := url.Parse(uri)
	if err != nil {
		res.Status = status.New(http.StatusInternalServerError, err)
		return
	}

	verb := ptr.Deref(opts.Verb, http.MethodGet)

	var body io.Reader
	if s := ptr.Deref(opts.Payload, ""); len(s) > 0 {
		body = strings.NewReader(s)
	}

	req, err := http.NewRequestWithContext(ctx, verb, u.String(), body)
	if err != nil {
		res.Status = status.New(http.StatusInternalServerError, err)
		return
	}

	if len(opts.Headers) > 0 {
		for _, el := range opts.Headers {
			idx := strings.Index(el, ":")
			if idx <= 0 {
				continue
			}
			req.Header.Set(el[:idx], el[idx+1:])
		}
	}

	cli, err := HTTPClientForEndpoint(opts.Endpoint)
	if err != nil {
		res.Status = status.New(http.StatusInternalServerError,
			fmt.Errorf("unable to create HTTP Client for endpoint: %w", err))
		return
	}

	respo, err := cli.Do(req)
	if err != nil {
		res.Status = status.New(http.StatusInternalServerError, err)
		return
	}
	defer respo.Body.Close()

	statusOK := respo.StatusCode >= 200 && respo.StatusCode < 300
	if !statusOK {
		dat, err := io.ReadAll(io.LimitReader(respo.Body, maxUnstructuredResponseTextBytes))
		if err != nil {
			res.Status = status.New(http.StatusInternalServerError, err)
			return
		}

		if err := json.Unmarshal(dat, &res.Status); err != nil {
			res.Status = status.New(res.Status.Code, fmt.Errorf("%s", string(dat)))
			return
		}

		return
	}

	dict, err := decodeResponseBody(respo)
	if err != nil {
		res.Status = status.New(http.StatusInternalServerError, err)
		return
	}

	res.Map = dict
	res.Status = status.New(http.StatusOK, nil)
	return
}

func decodeResponseBody(resp *http.Response) (map[string]any, error) {
	if !hasContentType(resp, "application/json") {
		return nil, fmt.Errorf("only 'application/json' media type is supported")
	}

	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	x := bytes.TrimSpace(dat)
	isArray := len(x) > 0 && x[0] == '['
	//isObject := len(x) > 0 && x[0] == '{'

	if isArray {
		v := []any{}
		err := json.Unmarshal(dat, &v)
		return map[string]any{
			"items": v,
		}, err
	}

	v := map[string]any{}
	err = json.Unmarshal(dat, &v)
	return v, err
}

// Determine whether the request `content-type` includes a
// server-acceptable mime-type
//
// Failure should yield an HTTP 415 (`http.StatusUnsupportedMediaType`)
func hasContentType(r *http.Response, mimetype string) bool {
	contentType := r.Header.Get("Content-type")
	if contentType == "" {
		return mimetype == "application/octet-stream"
	}

	for _, v := range strings.Split(contentType, ",") {
		t, _, err := mime.ParseMediaType(v)
		if err != nil {
			break
		}
		if t == mimetype {
			return true
		}
	}
	return false
}
