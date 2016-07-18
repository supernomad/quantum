package logger

import (
	"log"
	"os"
)

// Logger is a generic log interface
type Logger struct {
	Error        *log.Logger
	Debug        *log.Logger
	Info         *log.Logger
	Warn         *log.Logger
	DebugEnabled bool
}

func newLog(stream *os.File, level string) *log.Logger {
	return log.New(stream, level, log.LUTC|log.LstdFlags|log.Lshortfile)
}

// New Logger
func New(debugEnabled bool) *Logger {
	debug := newLog(os.Stdout, "[Debug] ")
	warn := newLog(os.Stderr, "[Warn] ")
	info := newLog(os.Stdout, "[Info] ")
	err := newLog(os.Stderr, "[Error] ")

	return &Logger{Error: err, Debug: debug, Info: info, Warn: warn, DebugEnabled: debugEnabled}
}
