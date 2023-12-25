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

var logger *log.Logger

func main() {
	logLevel, _ := env.GetOptionalString(env.LOCKSMITH_LOG_LEVEL, env.LOCKSMITH_LOG_LEVEL_DEFAULT)
	logger = log.New(log.Translate(logLevel))

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		signal_ch := make(chan os.Signal, 1)
		signal.Notify(signal_ch, syscall.SIGINT, syscall.SIGTERM)
		signal := <-signal_ch
		logger.Info("Got signal: ", signal)
		cancel()
	}()

	port, _ := env.GetOptionalUint16(env.LOCKSMITH_PORT, env.LOCKSMITH_PORT_DEFAULT)
	if err := server.New(&server.LocksmithOptions{Port: port}).Start(ctx); err != nil {
		logger.Error("Server start error: ", err)
		os.Exit(1)
	}

	logger.Info("Server stopped")
}
