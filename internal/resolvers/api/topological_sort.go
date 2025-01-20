package api

import (
	"fmt"

	templates "github.com/krateoplatformops/snowplow/apis/templates/v1"
	"github.com/krateoplatformops/snowplow/plumbing/ptr"
)

func topologicalSort(items []*templates.API) ([]string, error) {
	graph := make(map[string][]string)
	inDegree := make(map[string]int)
	itemSet := make(map[string]bool)

	for _, item := range items {
		itemSet[item.Name] = true

		if dep := ptr.Deref(item.DependOn, ""); dep != "" {
			graph[dep] = append(graph[dep], item.Name)
			inDegree[item.Name]++
		}
	}

	var queue []string
	for item := range itemSet {
		if inDegree[item] == 0 {
			queue = append(queue, item)
		}
	}

	var sortedItems []string
	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]
		sortedItems = append(sortedItems, item)

		for _, dependent := range graph[item] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	if len(sortedItems) != len(itemSet) {
		return nil, fmt.Errorf("cyclic dependency detected")
	}

	return sortedItems, nil
}
