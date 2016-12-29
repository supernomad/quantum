// Package common logger struct and func's
// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE
package common

import (
	"io/ioutil"
	"log"
	"os"
)

// Logger type to allow for a single point of logging configuration
type Logger struct {
	Error *log.Logger
	Info  *log.Logger
	Warn  *log.Logger
	Debug *log.Logger
}

// NewLogger creates a new logger based on user config.
func NewLogger(err, info, warn, debug bool) *Logger {
	logger := &Logger{
		Error: log.New(os.Stderr, "[ERROR]", 0),
		Info:  log.New(os.Stdout, "[INFO]", 0),
		Warn:  log.New(os.Stderr, "[WARN]", 0),
		Debug: log.New(os.Stdout, "[DEBUG]", 0),
	}

	if !err {
		logger.Error.SetOutput(ioutil.Discard)
	}

	if !info {
		logger.Info.SetOutput(ioutil.Discard)
	}

	if !warn {
		logger.Warn.SetOutput(ioutil.Discard)
	}

	if !debug {
		logger.Warn.SetOutput(ioutil.Discard)
	}

	return logger
}
