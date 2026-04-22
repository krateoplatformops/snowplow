package config

import "github.com/krateoplatformops/plumbing/env"

func authnNamespaceFromEnv() string {
	if val := env.String("AUTHN_NS", ""); val != "" {
		return val
	}

	return env.String("AUTHN_NAMESPACE", "")
}
