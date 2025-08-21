package util

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func ParseExtras(req *http.Request) (res map[string]any, err error) {
	res = map[string]any{}

	extrasParam := req.URL.Query().Get("extras")
	if extrasParam == "" {
		return
	}

	err = json.Unmarshal([]byte(extrasParam), &res)
	if err != nil {
		err = fmt.Errorf("invalid 'extras' parameter: %w", err)
		return
	}

	return
}
