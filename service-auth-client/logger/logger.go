package logger

import (
	"log"
	"os"
)

var (
	LogInfo  = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	LogError = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime)
	LogPanic = log.New(os.Stdout, "PANIC: ", log.Ldate|log.Ltime)
	LogFatal = log.New(os.Stdout, "FATAL: ", log.Ldate|log.Ltime)
)
