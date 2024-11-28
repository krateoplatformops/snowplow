package cmd

import (
	"context"
	"log/slog"

	"github.com/krateoplatformops/snowplow/internal/handlers/health"
	"github.com/krateoplatformops/snowplow/plumbing/server"
	"github.com/krateoplatformops/snowplow/plumbing/server/middlewares/clientconfig"
	"github.com/krateoplatformops/snowplow/plumbing/server/middlewares/cors"
	"github.com/krateoplatformops/snowplow/plumbing/server/middlewares/logger"
)

const (
	serviceName = "snowplow"
)

type Options struct {
	Log         *slog.Logger
	Build       string
	Debug       bool
	CorseOn     bool
	AuthnNS     string
	ServiceName string
	Port        int
}

func NewServer(ctx context.Context, opts Options) *server.Server {
	if len(opts.ServiceName) == 0 {
		opts.ServiceName = serviceName
	}

	srv := server.NewServer(opts.Port)

	if opts.CorseOn {
		srv.Use(cors.New(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{
				"Accept",
				"Authorization",
				"Content-Type",
				"X-Auth-Code",
				"X-Krateo-User",
				"X-Krateo-Groups",
			},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300,
		}).Handler)
	}

	srv.Use(logger.New(opts.Log))
	srv.Use(clientconfig.New(opts.AuthnNS, opts.Debug))

	// Register routes
	srv.Router.GET("/health", health.Check(opts.ServiceName, opts.Build))

	return srv
}
