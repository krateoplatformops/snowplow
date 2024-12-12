package actions

import (
	"strings"

	templates "github.com/krateoplatformops/snowplow/apis/templates/v1alpha1"
)

func mapVerbs(cat *templates.ActionTemplate) []string {
	verbs := []string{}
	x, ok := restToKube[strings.ToUpper(cat.Verb)]
	if ok {
		verbs = append(verbs, x)
		return verbs
	}

	for k := range kubeToREST {
		verbs = append(verbs, k)
	}
	return verbs
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
