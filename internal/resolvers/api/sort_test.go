package api

import (
	"fmt"
	"os"

	templates "github.com/krateoplatformops/snowplow/apis/templates/v1"
)

func Example_topologicalSort() {
	got, err := topologicalSort([]*templates.API{
		{Name: "api1", DependsOn: &templates.Dependency{Name: "api3"}},
		{Name: "api2"},
		{Name: "api3", DependsOn: &templates.Dependency{Name: "api2"}},
		{Name: "api4"},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		return
	}

	fmt.Println(got)

	// Output:
	// [api2 api4 api3 api1]
}
