package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/krateoplatformops/plumbing/env"
	"github.com/krateoplatformops/plumbing/server/probes"
	"github.com/krateoplatformops/plumbing/server/use"
	"github.com/krateoplatformops/plumbing/server/use/cors"
	_ "github.com/krateoplatformops/snowplow/docs"
	"github.com/krateoplatformops/snowplow/internal/config"
	"github.com/krateoplatformops/snowplow/internal/handlers"
	"github.com/krateoplatformops/snowplow/internal/handlers/dispatchers"
	"github.com/krateoplatformops/snowplow/internal/telemetry"
	httpSwagger "github.com/swaggo/http-swagger"
	"k8s.io/client-go/rest"
)

// @title SnowPlow API
// @version 0.1.0
// @description This the total new Krateo backend.
// @BasePath /
func main() {
	cfg := config.Setup()
	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	metrics, shutdownMetrics, err := telemetry.Setup(rootCtx, cfg.Log, telemetry.Config{
		Enabled:        cfg.OTelEnabled,
		ServiceName:    "snowplow",
		ExportInterval: cfg.OTelInterval,
	})
	if err != nil {
		cfg.Log.Error("OpenTelemetry setup failed", slog.Any("err", err))
		os.Exit(1)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownMetrics(ctx); err != nil {
			cfg.Log.Warn("OpenTelemetry shutdown failed", slog.Any("err", err))
		}
	}()

	server := newServer(cfg, metrics.WrapHTTP(newMux(cfg, metrics)))

	serverErr := make(chan error, 1)
	go func() {
		cfg.Log.Info("starting HTTP server", slog.Int("port", cfg.Port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	cfg.Log.Info("application is ready", slog.String("addr", server.Addr))
	metrics.IncStartupSuccess(rootCtx)

	select {
	case <-rootCtx.Done():
		cfg.Log.Info("shutdown signal received")
	case err := <-serverErr:
		metrics.IncStartupFailure(rootCtx)
		cfg.Log.Error("server error", slog.Any("err", err))
	}

	cfg.Log.Info("starting graceful shutdown")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(shutdownCtx); err != nil {
		cfg.Log.Error("HTTP server shutdown error", slog.Any("err", err))
		return
	}

	cfg.Log.Info("graceful shutdown complete")
}

func newMux(cfg *config.Config, metrics *telemetry.Metrics) *http.ServeMux {
	chain := use.NewChain(
		use.TraceId(),
		use.Logger(cfg.Log),
	)
	authChain := chain.Append(use.UserConfig(cfg.SigningKey, cfg.AuthnNS))

	mux := http.NewServeMux()

	// Register /livez and /readyz without auth — Kubernetes probes must not require JWT.
	probes.Register(mux, cfg.Log, clusterConfigPinger{}, time.Second)

	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)
	//mux.Handle("POST /convert", chain.Then(handlers.Converter()))

	mux.Handle("GET /api-info/names", chain.Then(handlers.Plurals()))
	mux.Handle("GET /list", authChain.Then(handlers.List(metrics)))

	mux.Handle("GET /call", authChain.Append(
		handlers.Dispatcher(dispatchers.All(cfg.AuthnNS))).
		Then(handlers.Call(cfg.Debug, metrics)))
	mux.Handle("POST /call", authChain.Then(handlers.Call(cfg.Debug, metrics)))
	mux.Handle("PUT /call", authChain.Then(handlers.Call(cfg.Debug, metrics)))
	mux.Handle("PATCH /call", authChain.Then(handlers.Call(cfg.Debug, metrics)))
	mux.Handle("DELETE /call", authChain.Then(handlers.Call(cfg.Debug, metrics)))

	mux.Handle("POST /jq", authChain.Then(handlers.JQ(metrics)))

	return mux
}

func newServer(cfg *config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr: ":" + strconv.Itoa(cfg.Port),
		Handler: use.CORS(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{
				"Accept",
				"Authorization",
				"Content-Type",
				"X-Auth-Code",
				"X-Krateo-TraceId",
			},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		})(handler),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 50 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
}

type clusterConfigPinger struct{}

func (clusterConfigPinger) Ping(context.Context) error {
	if env.TestMode() {
		return nil
	}

	_, err := rest.InClusterConfig()
	return err
}
