package kubeutil

import (
	"fmt"
	"os"
	"strings"
)

// ErrNoNamespace indicates that a namespace could not
// be found for the current environment
var (
	ErrNoNamespace = fmt.Errorf("namespace not found for current environment")
)

func ServiceAccountNamespace() (string, error) {
	nsBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNoNamespace
		}
		return "", err
	}

	ns := strings.TrimSpace(string(nsBytes))
	return ns, nil
}
