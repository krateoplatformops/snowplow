package middlewares

import (
	"net/http"

	"github.com/krateoplatformops/snowplow/plumbing/server/middlewares/cors"
)

func CORS(opts cors.Options) func(http.Handler) http.Handler {
	c := cors.New(opts)
	return c.Handler
}
