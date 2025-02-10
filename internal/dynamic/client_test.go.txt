//go:build integration
// +build integration

package dynamic_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/krateoplatformops/snowplow/internal/dynamic"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func TestList(t *testing.T) {
	rc, err := newRestConfig("")
	if err != nil {
		t.Fatal(err)
	}

	cli, err := dynamic.NewClient(rc)
	if err != nil {
		t.Fatal(err)
	}

	all, err := cli.Discover(context.TODO(), "defs")
	if err != nil {
		t.Fatal(err)
	}

	list := []unstructured.Unstructured{}
	for _, el := range all {
		obj, err := cli.List(context.TODO(), dynamic.Options{
			Namespace: "",
			GVR:       el,
		})
		if err != nil {
			t.Fatal(err)
		}

		for _, x := range obj.Items {
			unstructured.RemoveNestedField(
				x.UnstructuredContent(), "metadata", "managedFields")
			list = append(list, x)
		}
	}

	dat, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(dat))
}

func TestDiscover(t *testing.T) {
	filter := "compositions"

	rc, err := newRestConfig("")
	if err != nil {
		t.Fatal(err)
	}

	cli, err := dynamic.NewClient(rc)
	if err != nil {
		t.Fatal(err)
	}

	all, err := cli.Discover(context.TODO(), filter)
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(all)
}

func newRestConfig(fn string) (*rest.Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	if len(fn) == 0 {
		fn = filepath.Join(home, ".kube", "config")
	} else {
		fn = filepath.Join(home, fn)
	}

	return clientcmd.BuildConfigFromFlags("", fn)
}
