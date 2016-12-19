// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE
package common

import (
	"log"
	"os"
)

// Logger type to allow for a single point of logging configuration
type Logger struct {
	Info  *log.Logger
	Error *log.Logger
}

// NewLogger creates a new logger based on user config.
func NewLogger() *Logger {
	info := log.New(os.Stdout, "[INFO]", 0)
	err := log.New(os.Stderr, "[ERROR]", 0)

	return &Logger{
		Info:  info,
		Error: err,
	}
}
