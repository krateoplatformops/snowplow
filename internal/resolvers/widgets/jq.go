package widgets

import (
	"fmt"
	"sort"

	"github.com/krateoplatformops/snowplow/plumbing/maps"
)

func evalJQ(in map[string]any, ds map[string]any) error {
	if len(ds) == 0 || len(in) == 0 {
		return nil
	}

	paths := maps.LeafPaths(in, "")
	sort.Strings(paths)

	for _, path := range paths {
		fields := maps.ParsePath(path)

		value, found := maps.NestedValue(in, fields)
		if !found {
			continue
		}

		fmt.Printf("Path: %s, Value: %v\n", path, value)
		// if the value is string, we can try to evaluate a JQ expression
		if strValue, ok := value.(string); ok {
			fmt.Printf("  ==> maybe evaluate JQ: %s\n", strValue)
		}
	}

	return nil
}
