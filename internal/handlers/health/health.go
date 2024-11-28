package health

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/krateoplatformops/snowplow/plumbing/http/response/status"
	"github.com/krateoplatformops/snowplow/plumbing/server"
	"k8s.io/client-go/rest"
)

// @Summary Liveness Endpoint
// @Description Health Check
// @ID health
// @Produce  json
// @Success 200 {object} ServiceInfo
// @Router /health [get]
func Check(serviceName, build string) server.Handler {
	return func(wri http.ResponseWriter, req *http.Request) {
		if _, err := rest.InClusterConfig(); err != nil {
			status.ServiceUnavailable(wri, err)
			return
		}

		ns, _ := getServiceAccountNamespace()

		data := ServiceInfo{
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

type ServiceInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Build     string `json:"build"`
}

// errNoNamespace indicates that a namespace could not
// be found for the current environment
var (
	errNoNamespace = fmt.Errorf("namespace not found for current environment")
)

func getServiceAccountNamespace() (string, error) {
	nsBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		if os.IsNotExist(err) {
			return "", errNoNamespace
		}
		return "", err
	}

	ns := strings.TrimSpace(string(nsBytes))
	return ns, nil
}
