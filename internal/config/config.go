package config

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/krateoplatformops/plumbing/env"
	"github.com/krateoplatformops/plumbing/logger"
	jqsupport "github.com/krateoplatformops/snowplow/internal/support/jq"
)

const (
	serviceName               = "snowplow"
	defaultDebug              = false
	defaultOtelEnabled        = false
	defaultOtelExportInterval = 30 * time.Second
)

type Config struct {
	Port          int
	Debug         bool
	Blizzard      bool
	SigningKey    string
	AuthnNS       string
	JQModulesPath string
	OTelEnabled   bool
	OTelInterval  time.Duration
	Log           *slog.Logger
}

func Setup() *Config {
	cfg := &Config{}

	cfgPort := flag.Int("port",
		env.ServicePort("PORT", 8081),
		"port to listen on",
	)

	cfgDebug := flag.Bool("debug",
		env.Bool("DEBUG", defaultDebug),
		"enable or disable debug logs",
	)

	cfgBlizzard := flag.Bool("blizzard",
		env.Bool("BLIZZARD", false),
		"dump verbose output",
	)

	cfgAuthnNS := flag.String("authn-namespace",
		authnNamespaceFromEnv(),
		"krateo authn service clientconfig secrets namespace",
	)

	cfgJWTSignKey := flag.String("jwt-sign-key",
		env.String("JWT_SIGN_KEY", ""),
		"secret key used to sign JWT tokens",
	)

	cfgJQModulesPath := flag.String("jq-modules-path",
		env.String(jqsupport.EnvModulesPath, ""),
		"loads JQ custom modules from the filesystem",
	)

	cfgOTelEnabled := flag.Bool("otel-enabled",
		env.Bool("OTEL_ENABLED", defaultOtelEnabled),
		"enable OpenTelemetry metrics exporter",
	)

	cfgOTelExportInterval := flag.Duration("otel-export-interval",
		env.Duration("OTEL_EXPORT_INTERVAL", defaultOtelExportInterval),
		"OpenTelemetry metric export interval",
	)

	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Flags:")
		flag.PrintDefaults()
	}

	flag.Parse()

	cfg.Port = *cfgPort
	cfg.Debug = *cfgDebug
	cfg.Blizzard = *cfgBlizzard
	cfg.AuthnNS = *cfgAuthnNS
	cfg.SigningKey = *cfgJWTSignKey
	cfg.JQModulesPath = *cfgJQModulesPath
	cfg.OTelEnabled = *cfgOTelEnabled
	cfg.OTelInterval = *cfgOTelExportInterval
	cfg.Log = logger.New(serviceName, cfg.Debug)

	cfg.syncLegacyEnv()

	return cfg
}

func (c *Config) syncLegacyEnv() {
	os.Setenv("DEBUG", fmt.Sprintf("%t", c.Debug))
	os.Setenv("TRACE", fmt.Sprintf("%t", c.Blizzard))
	os.Setenv("AUTHN_NAMESPACE", c.AuthnNS)
	os.Setenv("AUTHN_NS", c.AuthnNS)
	os.Setenv(jqsupport.EnvModulesPath, c.JQModulesPath)
}
