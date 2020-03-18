package logger

import (
	"log"
	"os"

	Args "github.com/arush15june/load-tester/src/tester/args"
)

// Logging Wrappers
var (
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lmicroseconds)
	LogInfo    = func(data string) {
		InfoLogger.Output(2, data)
	}

	WarnLogger = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lmicroseconds)
	LogWarn    = func(data string) {
		WarnLogger.Output(2, data)
	}

	FatalLogger = log.New(os.Stdout, "FATAL: ", log.Ldate|log.Ltime|log.Lmicroseconds)
	LogFatal    = func(data string) {
		FatalLogger.Output(2, data)
	}

	DebugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lmicroseconds)
	LogDebug    = func(data string) {
		if *Args.Verbose {
			DebugLogger.Output(2, data)
		}
	}
)
