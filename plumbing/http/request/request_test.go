//go:build integration
// +build integration

package request_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/krateoplatformops/snowplow/plumbing/endpoints"
	httpcall "github.com/krateoplatformops/snowplow/plumbing/http/request"
	"github.com/krateoplatformops/snowplow/plumbing/ptr"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func TestCallNoProxy(t *testing.T) {
	rc, err := newRestConfig()
	if err != nil {
		t.Fatal(err)
	}

	authn, err := endpoints.FromSecret(context.TODO(), rc,
		"cyberjoker-clientconfig", "default")
	if err != nil {
		t.Fatal(err)
	}

	res := httpcall.Do(context.TODO(), httpcall.Options{
		Path: ptr.To("/anything"),
		Verb: ptr.To("POST"),
		Headers: []string{
			"User-Agent: Krateo",
			"X-Data-1: XXXXXX",
			"X-Data-2: YYYYYY",
		},
		Payload:  ptr.To(`{"name": "John", "surname": "Doe"}`),
		Endpoint: authn,
	})

	spew.Dump(res)
}

func newRestConfig() (*rest.Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return clientcmd.BuildConfigFromFlags("", filepath.Join(home, ".kube", "config"))
}
