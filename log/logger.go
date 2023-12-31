package log

import (
	"log"
	"os"
)

type LogLevel uint8

// Global logger used for direct calls to Debug, Info, Warning, Error, and Fatal.
// Defaults to the WARNING log level, call setLogLevel to overwrite.
var globalLogger *Logger = New(DEBUG)

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

	if logLevel == DEBUG {
		logger.debug = log.New(os.Stdout, "DEBUG   ", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
	}
	if logLevel <= INFO {
		logger.info = log.New(os.Stdout, "INFO    ", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
	}
	if logLevel <= WARNING {
		logger.warning = log.New(os.Stdout, "WARNING ", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
	}
	if logLevel <= ERROR {
		logger.err = log.New(os.Stdout, "ERROR   ", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
	}
	if logLevel <= FATAL {
		logger.fatal = log.New(os.Stderr, "FATAL   ", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
	}

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
		logger.debug.Println(args...)
	}
}

func (logger *Logger) Info(args ...any) {
	if logger.logLevel <= INFO {
		logger.info.Println(args...)
	}
}

func (logger *Logger) Warning(args ...any) {
	if logger.logLevel <= WARNING {
		logger.warning.Println(args...)
	}
}

func (logger *Logger) Error(args ...any) {
	if logger.logLevel <= ERROR {
		logger.err.Println(args...)
	}
}

func (logger *Logger) Fatal(args ...any) {
	logger.fatal.Fatalln(args...)
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
