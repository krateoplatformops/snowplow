package use

import (
	"net/http"

	"github.com/krateoplatformops/snowplow/plumbing/server/use/cors"
)

func CORS(opts cors.Options) func(http.Handler) http.Handler {
	c := cors.New(opts)
	return c.Handler
}
