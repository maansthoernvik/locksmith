package main

import (
	"github.com/maansthoernvik/locksmith/env"
	"github.com/maansthoernvik/locksmith/log"
	"github.com/maansthoernvik/locksmith/server"
)

func main() {
	logLevel, _ := env.GetRequiredString("LOCKSMITH_LOG_LEVEL")

	switch logLevel {
	case "DEBUG":
		log.GlobalLogger = log.New(log.DEBUG)
		break
	case "INFO":
		log.GlobalLogger = log.New(log.INFO)
		break
	case "WARNING":
		log.GlobalLogger = log.New(log.WARNING)
		break
	case "ERROR":
		log.GlobalLogger = log.New(log.ERROR)
		break
	case "FATAL":
		log.GlobalLogger = log.New(log.FATAL)
		break
	default:
		break
	}

	port, _ := env.GetOptionalUint16(env.LOCKSMITH_PORT, env.LOCKSMITH_PORT_DEFAULT)
	if err := server.New().Start(port); err != nil {
		log.GlobalLogger.Error("Server start error: ", err)
	}
	log.GlobalLogger.Info("Server started on port: ", port)
}
