package main

import (
	"github.com/maansthoernvik/locksmith/log"
)

func main() {
	logger := log.New(log.DEBUG)

	logger.Debug("this wont show")
	logger.Info("this INFO shows")
	logger.Warning("this warning shows")
	logger.Error("this error shows")
	logger.Fatal("this fatal error causes a crash")
}
