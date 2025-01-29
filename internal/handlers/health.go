package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/krateoplatformops/snowplow/plumbing/http/response"
	"github.com/krateoplatformops/snowplow/plumbing/kubeutil"
	"k8s.io/client-go/rest"
)

// @Summary Liveness Endpoint
// @Description Health HealthCheck
// @ID health
// @Produce  json
// @Success 200 {object} serviceInfo
// @Router /health [get]
func HealthCheck(serviceName, build string) http.HandlerFunc {
	return func(wri http.ResponseWriter, req *http.Request) {
		if _, err := rest.InClusterConfig(); err != nil {
			response.ServiceUnavailable(wri, err)
			return
		}

		ns, _ := kubeutil.ServiceAccountNamespace()

		data := serviceInfo{
			Name:      serviceName,
			Build:     build,
			Namespace: ns,
		}

		wri.Header().Set("Content-Type", "application/json")
		wri.WriteHeader(http.StatusOK)
		json.NewEncoder(wri).Encode(data)
		return
	}
}

type serviceInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Build     string `json:"build"`
}
