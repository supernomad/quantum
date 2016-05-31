package logger

import (
	"log"
	"os"
)

type Logger struct {
	error        *log.Logger
	debug        *log.Logger
	info         *log.Logger
	warn         *log.Logger
	debugEnabled bool
}

func (l *Logger) Info(v ...interface{}) {
	l.info.Println(v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.error.Println(v...)
}

func (l *Logger) Debug(v ...interface{}) {
	if l.debugEnabled {
		l.debug.Println(v...)
	}
}

func (l *Logger) Warn(v ...interface{}) {
	l.warn.Println(v...)
}

func New(debugEnabled bool) *Logger {
	err := log.New(os.Stderr,
		"[Error]",
		log.Ldate|log.Ltime|log.Lshortfile)
	info := log.New(os.Stdout,
		"[INFO]",
		log.Ldate|log.Ltime|log.Lshortfile)
	debug := log.New(os.Stdout,
		"[DEBUG]",
		log.Ldate|log.Ltime|log.Lshortfile)
	warn := log.New(os.Stdout,
		"[WARN]",
		log.Ldate|log.Ltime|log.Lshortfile)

	return &Logger{error: err, debug: debug, info: info, warn: warn, debugEnabled: debugEnabled}
}
