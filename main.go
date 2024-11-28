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

	"github.com/krateoplatformops/snowplow/cmd"
	"github.com/krateoplatformops/snowplow/plumbing/env"
)

var (
	Build string
)

// @title Backend API
// @version 1.0
// @description This the Krateo BFF server.
// @BasePath /
func main() {
	opts := cmd.Options{
		Build: Build,
	}
	flag.BoolVar(&opts.Debug, "debug", env.Bool("DEBUG", false), "dump verbose output")
	flag.BoolVar(&opts.CorseOn, "cors", env.Bool("CORS", true), "enable or disable CORS")
	flag.IntVar(&opts.Port, "port", env.ServicePort("PORT", 8080), "port to listen on")
	flag.StringVar(&opts.AuthnNS, "authn-store-namespace",
		env.String("AUTHN_STORE_NAMESPACE", ""),
		"krateo authn service clientconfig secrets namespace")

	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Flags:")
		flag.PrintDefaults()
	}

	flag.Parse()

	logLevel := slog.LevelInfo
	if opts.Debug {
		logLevel = slog.LevelDebug
	}
	opts.Log = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}))

	srv := cmd.NewServer(context.Background(), opts)

	ctx, stop := signal.NotifyContext(context.Background(), []os.Signal{
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGKILL,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	}...)
	defer stop()

	go func() {
		if err := srv.Run(); err != nil && err != http.ErrServerClosed {
			opts.Log.Error("server cannot run",
				slog.Int("port", opts.Port),
				slog.Any("err", err))
		}
	}()

	// Listen for the interrupt signal.
	opts.Log.Info("server is ready to handle requests", slog.Int("port", opts.Port))
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	opts.Log.Info("server is shutting down gracefully, press Ctrl+C again to force")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		opts.Log.Error("server forced to shutdown", slog.Any("err", err))
	}

	opts.Log.Info("server gracefully stopped")
}
