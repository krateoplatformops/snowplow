package actions

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/krateoplatformops/snowplow/apis/templates/v1alpha1"
)

func TestMapVerbs(t *testing.T) {
	table := []struct {
		in  string
		out []string
	}{
		{"post", []string{"create"}},
		{"Put", []string{"update"}},
		{"gEt", []string{"get"}},
		{"get", []string{"get"}},
		{"", []string{"create", "update", "delete", "get"}},
	}

	for _, tc := range table {
		got := mapVerbs(&v1alpha1.ActionTemplate{
			Verb: tc.in,
		})

		if diff := cmp.Diff(got, tc.out); len(diff) > 0 {
			t.Fatal(diff)
		}
	}
}
