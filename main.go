package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/maansthoernvik/locksmith/env"
	"github.com/maansthoernvik/locksmith/log"
	"github.com/maansthoernvik/locksmith/server"
)

func main() {
	logLevel, _ := env.GetRequiredString("LOCKSMITH_LOG_LEVEL")

	switch logLevel {
	case "DEBUG":
		log.GlobalLogger = log.New(log.DEBUG)
	case "INFO":
		log.GlobalLogger = log.New(log.INFO)
	case "WARNING":
		log.GlobalLogger = log.New(log.WARNING)
	case "ERROR":
		log.GlobalLogger = log.New(log.ERROR)
	case "FATAL":
		log.GlobalLogger = log.New(log.FATAL)
	default:
		break
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		signal_ch := make(chan os.Signal, 1)
		signal.Notify(signal_ch, syscall.SIGINT, syscall.SIGTERM)
		signal := <-signal_ch
		log.GlobalLogger.Info("Got signal: ", signal)
		cancel()
	}()

	port, _ := env.GetOptionalUint16(env.LOCKSMITH_PORT, env.LOCKSMITH_PORT_DEFAULT)
	if err := server.New(&server.LocksmithOptions{Port: port}).Start(ctx); err != nil {
		log.GlobalLogger.Error("Server start error: ", err)
		os.Exit(1)
	}

	log.GlobalLogger.Info("Server stopped")
}
