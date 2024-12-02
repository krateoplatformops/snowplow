package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/krateoplatformops/snowplow/internal/handlers/health"
	"github.com/krateoplatformops/snowplow/internal/handlers/resources"
	"github.com/krateoplatformops/snowplow/plumbing/env"
	"github.com/krateoplatformops/snowplow/plumbing/server/middlewares"
	"github.com/krateoplatformops/snowplow/plumbing/server/middlewares/cors"

	_ "github.com/krateoplatformops/snowplow/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

const (
	serviceName = "snowplow"
)

var (
	build string
)

// @title SnowPlow API
// @version 0.1.0
// @description This the total new Krateo backend.
// @BasePath /
func main() {
	debugOn := flag.Bool("debug", env.Bool("DEBUG", false), "enable or disable debug logs")
	corsOn := flag.Bool("cors", env.Bool("CORS", true), "enable or disable CORS")
	port := flag.Int("port", env.ServicePort("PORT", 8081), "port to listen on")
	authnNS := flag.String("authn-store-namespace",
		env.String("AUTHN_STORE_NAMESPACE", ""),
		"krateo authn service clientconfig secrets namespace")
	blizzard := flag.Bool("blizzard", env.Bool("BLIZZARD", false), "dump verbose output")

	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Flags:")
		flag.PrintDefaults()
	}

	flag.Parse()

	logLevel := slog.LevelInfo
	if *debugOn {
		logLevel = slog.LevelDebug
	}
	log := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}))

	if *debugOn {
		if *blizzard {
			os.Setenv("BLIZZARD", "true")
		}

		log.Debug("environment variables", slog.Any("env", os.Environ()))
	}

	var base middlewares.Chain
	if *corsOn {
		base = base.Append(middlewares.CORS(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{
				"Accept",
				"Authorization",
				"Content-Type",
				"X-Auth-Code",
				"X-Krateo-TraceId",
				"X-Krateo-User",
				"X-Krateo-Groups",
			},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}))
	}
	base = base.Append(middlewares.TraceId(), middlewares.Logger(log))

	mux := http.NewServeMux()
	mux.Handle("GET /health",
		health.Check(serviceName, build))
	mux.Handle("GET /swagger/",
		httpSwagger.WrapHandler)
	mux.Handle("GET /list",
		base.Append(middlewares.RESTConfig(*authnNS, *debugOn)).
			Then(resources.List(*authnNS, *debugOn)))

	//mux.Handle("GET /call", call.Call(*authnNS, *debugOn))
	//mux.Handle("POST /call", call.Call(*authnNS, *debugOn))
	//mux.Handle("PUT /call", call.Call(*authnNS, *debugOn))
	//mux.Handle("DELETE /call", call.Call(*authnNS, *debugOn))
	//mux.Handle("GET /api-info/names", info.Names())

	ctx, stop := signal.NotifyContext(context.Background(), []os.Signal{
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGKILL,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	}...)
	defer stop()

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 50 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server cannot run",
				slog.String("addr", server.Addr),
				slog.Any("err", err))
		}
	}()

	// Listen for the interrupt signal.
	log.Info("server is ready to handle requests", slog.String("addr", server.Addr))
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Info("server is shutting down gracefully, press Ctrl+C again to force")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", slog.Any("err", err))
	}

	log.Info("server gracefully stopped")
}
