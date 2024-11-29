package kubeconfig

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/krateoplatformops/snowplow/plumbing/shortid"
)

func Transport(next http.RoundTripper) http.RoundTripper {
	return RoundTripFunc(func(in *http.Request) (resp *http.Response, err error) {
		req := cloneRequest(in)

		traceID, _ := req.Context().Value(requestIdContextKey).(string)
		if traceID == "" {
			traceID = shortid.MustGenerate()
		}
		req.Header.Set(string(requestIdContextKey), traceID)

		// Dump the request to os.Stderr.
		b, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return nil, err
		}
		os.Stderr.Write(b)
		os.Stderr.Write([]byte{'\n'})

		resp, err = next.RoundTrip(req)
		// If an error was returned, dump it to os.Stderr.
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return resp, err
		}

		// Dump the response to os.Stderr.
		b, err = httputil.DumpResponse(resp, req.URL.Query().Get("watch") != "true")
		if err != nil {
			return nil, err
		}
		os.Stderr.Write(b)
		os.Stderr.Write([]byte{'\n'})

		return resp, err
	})
}

// RoundTripFunc, similar to http.HandlerFunc, is an adapter
// to allow the use of ordinary functions as http.RoundTrippers.
type RoundTripFunc func(r *http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// cloneRequest creates a shallow copy of a given request
// to comply with stdlib's http.RoundTripper contract:
//
// RoundTrip should not modify the request, except for
// consuming and closing the Request's Body. RoundTrip may
// read fields of the request in a separate goroutine. Callers
// should not mutate or reuse the request until the Response's
// Body has been closed.
func cloneRequest(orig *http.Request) *http.Request {
	clone := &http.Request{}
	*clone = *orig

	clone.Header = make(http.Header, len(orig.Header))
	for key, value := range orig.Header {
		clone.Header[key] = append([]string{}, value...)
	}

	return clone
}

type contextKey string

const (
	requestIdContextKey contextKey = "X-Request-Id"
)
