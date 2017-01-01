// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package common

import (
	"io/ioutil"
	"log"
	"os"
)

const (
	// NoopLogger will noop all logging calls this is only used for testing
	NoopLogger = iota
	// ErrorLogger will only output error logs
	ErrorLogger
	// WarnLogger will output warn/error logs
	WarnLogger
	// InfoLogger will output info/warn/error logs
	InfoLogger
	// DebugLogger will output debug/info/warn/error logs
	DebugLogger
)

// Logger struct to allow for a single point for logging configuration
type Logger struct {
	Error *log.Logger
	Info  *log.Logger
	Warn  *log.Logger
	Debug *log.Logger
}

// NewLogger creates a new logger struct based on user config.
func NewLogger(kind int) *Logger {
	logger := &Logger{
		Error: log.New(os.Stderr, "[ERROR] ", 0),
		Warn:  log.New(os.Stderr, "[WARN] ", 0),
		Info:  log.New(os.Stdout, "[INFO] ", 0),
		Debug: log.New(os.Stdout, "[DEBUG] ", 0),
	}

	if ErrorLogger > kind {
		logger.Error.SetOutput(ioutil.Discard)
	}

	if WarnLogger > kind {
		logger.Warn.SetOutput(ioutil.Discard)
	}

	if InfoLogger > kind {
		logger.Info.SetOutput(ioutil.Discard)
	}

	if DebugLogger > kind {
		logger.Warn.SetOutput(ioutil.Discard)
	}

	return logger
}
