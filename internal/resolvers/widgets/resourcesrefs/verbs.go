package resourcesrefs

import (
	"strings"
)

func mapVerbs(verb string) []string {
	all := []string{}
	x, ok := restToKube[strings.ToUpper(verb)]
	if ok {
		all = append(all, x)
		return all
	}

	for k := range kubeToREST {
		if !contains(all, k) {
			all = append(all, k)
		}
	}

	return all
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

var (
	kubeToREST = map[string]string{
		"create": "POST",
		"update": "PUT",
		"delete": "DELETE",
		"get":    "GET",
	}

	restToKube = map[string]string{
		"POST":   "create",
		"PUT":    "update",
		"DELETE": "delete",
		"GET":    "get",
	}
)
