package log

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger zap.Logger

func init() {
	config := zap.NewProductionConfig()

	// TODO: Settings should be based on environment variables
	config.Level.SetLevel(zap.DebugLevel)
	config.Encoding = "console"
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	config.EncoderConfig.ConsoleSeparator = " "

	logger, err := config.Build()
	Logger = *logger
	if err != nil {
		log.Fatal("Failed to initialize zap logger:", err)
	}
}
