package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	locksmith "github.com/maansthoernvik/locksmith/pkg"
	"github.com/maansthoernvik/locksmith/pkg/env"
	"github.com/maansthoernvik/locksmith/pkg/vault"
	"github.com/maansthoernvik/locksmith/pkg/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Set global log level
	logLevel, _ := env.GetOptionalString(env.LOCKSMITH_LOG_LEVEL, env.LOCKSMITH_LOG_LEVEL_DEFAULT)
	zerolog.SetGlobalLevel(translateToZerologLevel(logLevel))
	if console, _ := env.GetOptionalBool(env.LOCKSMITH_LOG_OUTPUT_CONSOLE, env.LOCKSMITH_LOG_OUTPUT_CONSOLE_DEFAULT); console {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Print to bypass loglevel settings and write to stdout
	// Check if '?' since the version info can only be set for container builds, not via 'go install'
	if version.Version != "?" {
		log.Info().
			Str("version", version.Version).
			Str("commit", version.Commit).
			Str("built", version.Built).
			Msg("starting Locksmith")
	} else {
		log.Info().Msg("starting locksmith")
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Check if Prometheus metrics are enabled
	var srv *http.Server
	if metrics, _ := env.GetOptionalBool(env.LOCKSMITH_METRICS, env.LOCKSMITH_METRICS_DEFAULT); metrics {
		http.Handle("/metrics", promhttp.Handler())
		srv = &http.Server{Addr: ":20000"}
		go func() {
			log.Info().Str("address", srv.Addr).Msg("starting metrics server")
			if err := srv.ListenAndServe(); err != http.ErrServerClosed {
				log.Error().Err(err).Msg("metrics server failure")
			} else {
				log.Info().Msg("stopped metrics server")
			}
		}()
	}

	go func() {
		signal_ch := make(chan os.Signal, 1)
		signal.Notify(signal_ch, syscall.SIGINT, syscall.SIGTERM)
		signal := <-signal_ch
		log.Info().Any("signal", signal).Msg("captured stop signal")
		if err := srv.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("error shutting down metrics server")
		}
		cancel()
	}()

	port, _ := env.GetOptionalUint16(env.LOCKSMITH_PORT, env.LOCKSMITH_PORT_DEFAULT)
	queueType, _ := env.GetOptionalString(env.LOCKSMITH_Q_TYPE, env.LOCKSMITH_Q_TYPE_DEFAULT)
	concurrency, _ := env.GetOptionalInteger(env.LOCKSMITH_Q_CONCURRENCY, env.LOCKSMITH_Q_CONCURRENCY_DEFAULT)
	capacity, _ := env.GetOptionalInteger(env.LOCKSMITH_Q_CAPACITY, env.LOCKSMITH_Q_CAPACITY_DEFAULT)

	locksmithOptions := &locksmith.LocksmithOptions{
		Port:             port,
		QueueType:        vault.QueueType(queueType),
		QueueConcurrency: concurrency,
		QueueCapacity:    capacity,
	}
	if tls, _ := env.GetOptionalBool(env.LOCKSMITH_TLS, env.LOCKSMITH_TLS_DEFAULT); tls {
		locksmithOptions.TlsConfig = getTlsConfig()
	}
	if err := locksmith.New(locksmithOptions).Start(ctx); err != nil {
		log.Error().Err(err).Msg("server start error")
		os.Exit(1)
	}

	log.Info().Msg("server stopped")
}

func translateToZerologLevel(level string) zerolog.Level {
	switch level {
	case "DEBUG":
		return zerolog.DebugLevel
	case "INFO":
		return zerolog.InfoLevel
	case "WARNING":
		return zerolog.WarnLevel
	case "ERROR":
		return zerolog.ErrorLevel
	case "FATAL":
		return zerolog.FatalLevel
	case "PANIC":
		return zerolog.PanicLevel
	}

	log.Warn().Msg("unable to decode log level")
	return zerolog.NoLevel
}

// Fetch TLS config to supply the TCP listener.
func getTlsConfig() *tls.Config {
	tlsConfig := &tls.Config{}

	serverCertPath, _ := env.GetOptionalString(env.LOCKSMITH_TLS_CERT_PATH, env.LOCKSMITH_TLS_CERT_PATH_DEFAULT)
	serverKeyPath, _ := env.GetOptionalString(env.LOCKSMITH_TLS_KEY_PATH, env.LOCKSMITH_TLS_KEY_PATH_DEFAULT)
	cert, err := tls.LoadX509KeyPair(serverCertPath, serverKeyPath)
	if err != nil {
		panic("failed to load server cert/key pair")
	}
	tlsConfig.Certificates = []tls.Certificate{cert}

	requireClientVerify, _ := env.GetOptionalBool(env.LOCKSMITH_TLS_REQUIRE_CLIENT_CERT, env.LOCKSMITH_TLS_REQUIRE_CLIENT_CERT_DEFAULT)
	if requireClientVerify {
		clientCaCertPath, _ := env.GetOptionalString(env.LOCKSMITH_TLS_CLIENT_CA_CERT_PATH, env.LOCKSMITH_TLS_CLIENT_CA_CERT_PATH_DEFAULT)
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		caCert, err := os.ReadFile(clientCaCertPath)
		if err != nil {
			panic("failed to read client CA cert file")
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(caCert)
		tlsConfig.ClientCAs = pool
	}

	return tlsConfig
}
