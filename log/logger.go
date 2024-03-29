// Package log wraps Golang's stdlib logging to allow for setting a log level.
package log

import (
	"fmt"
	"log"
	"os"

	"github.com/maansthoernvik/locksmith/env"
)

// Global logger used for direct calls to Debug, Info, Warning, Error, and Fatal.
// Defaults to the DEBUG log level, call SetLogLevel to overwrite.
var globalLogger *Logger = New(DEBUG)

func init() {
	val, _ := env.GetOptionalString(env.LOCKSMITH_LOG_LEVEL, env.LOCKSMITH_LOG_LEVEL_DEFAULT)
	globalLogger = New(Translate(val))
}

// Calldepth to capture callers rather than logger.go
var calldepth = 3

type LogLevel uint8

const (
	DEBUG   LogLevel = 0
	INFO    LogLevel = 1
	WARNING LogLevel = 2
	ERROR   LogLevel = 3
	FATAL   LogLevel = 4
)

type Logger struct {
	logLevel LogLevel
	debug    *log.Logger
	info     *log.Logger
	warning  *log.Logger
	err      *log.Logger
	fatal    *log.Logger
}

func New(logLevel LogLevel) *Logger {
	var logger = &Logger{logLevel: logLevel}

	logger.debug = log.New(os.Stdout, "DEBUG   ", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
	logger.info = log.New(os.Stdout, "INFO    ", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
	logger.warning = log.New(os.Stderr, "WARNING ", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
	logger.err = log.New(os.Stderr, "ERROR   ", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
	logger.fatal = log.New(os.Stderr, "FATAL   ", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)

	return logger
}

func Translate(logStr string) LogLevel {
	switch logStr {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARNING":
		return WARNING
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return WARNING
	}
}

func (logger *Logger) Debug(args ...any) {
	if logger.logLevel == DEBUG {
		_ = logger.debug.Output(calldepth, fmt.Sprint(args...))
	}
}

func (logger *Logger) Info(args ...any) {
	if logger.logLevel <= INFO {
		_ = logger.info.Output(calldepth, fmt.Sprint(args...))
	}
}

func (logger *Logger) Warning(args ...any) {
	if logger.logLevel <= WARNING {
		_ = logger.warning.Output(calldepth, fmt.Sprint(args...))
	}
}

func (logger *Logger) Error(args ...any) {
	if logger.logLevel <= ERROR {
		_ = logger.err.Output(calldepth, fmt.Sprint(args...))
	}
}

func (logger *Logger) Fatal(args ...any) {
	_ = logger.fatal.Output(calldepth, fmt.Sprint(args...))
	os.Exit(1)
}

func Debug(args ...any) {
	globalLogger.Debug(args...)
}

func Info(args ...any) {
	globalLogger.Info(args...)
}

func Warning(args ...any) {
	globalLogger.Warning(args...)
}

func Error(args ...any) {
	globalLogger.Error(args...)
}

func Fatal(args ...any) {
	globalLogger.Fatal(args...)
}

func SetLogLevel(logLevel LogLevel) {
	globalLogger.logLevel = logLevel
}
