// Copyright (c) 2016-2018 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package common

import (
	"io/ioutil"
	"log"
	"os"
)

// LoggerType will determine the logging level of the logger object created.
type LoggerType int

const (
	// NoopLogger will noop all logging calls this is only used for testing.
	NoopLogger LoggerType = iota

	// ErrorLogger will only output error logs.
	ErrorLogger

	// WarnLogger will output warn/error logs.
	WarnLogger

	// InfoLogger will output info/warn/error logs.
	InfoLogger

	// DebugLogger will output debug/info/warn/error logs.
	DebugLogger
)

// Logger struct which allows for a single global point for logging configuration.
type Logger struct {
	Plain *log.Logger
	Error *log.Logger
	Info  *log.Logger
	Warn  *log.Logger
	Debug *log.Logger
}

// NewLogger creates a new logger struct based on the supplied LoggerType.
func NewLogger(loggerType LoggerType) *Logger {
	logger := &Logger{
		Plain: log.New(os.Stdout, "", 0),
		Error: log.New(os.Stderr, "[ERROR] ", 0),
		Warn:  log.New(os.Stderr, "[WARN] ", 0),
		Info:  log.New(os.Stdout, "[INFO] ", 0),
		Debug: log.New(os.Stdout, "[DEBUG] ", 0),
	}

	if ErrorLogger > loggerType {
		logger.Error.SetOutput(ioutil.Discard)
	}

	if WarnLogger > loggerType {
		logger.Warn.SetOutput(ioutil.Discard)
	}

	if InfoLogger > loggerType {
		logger.Info.SetOutput(ioutil.Discard)
		logger.Plain.SetOutput(ioutil.Discard)
	}

	if DebugLogger > loggerType {
		logger.Debug.SetOutput(ioutil.Discard)
	}

	return logger
}
