package backendendpoints

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
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
		{"", []string{"create", "delete", "get", "update"}},
	}

	for _, tc := range table {
		got := mapVerbs(tc.in)
		sort.Strings(got)

		if diff := cmp.Diff(got, tc.out); len(diff) > 0 {
			t.Fatal(diff)
		}
	}
}
