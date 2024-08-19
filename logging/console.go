package logger

import (
	"fmt"
	"log"
	"os"
)

type ConsoleLogger struct {
	debug    *log.Logger
	info     *log.Logger
	warning  *log.Logger
	error    *log.Logger
	critical *log.Logger
	level    int
}

func NewConsoleLogger(name string, level int) *ConsoleLogger {
	cl := &ConsoleLogger{
		debug:    log.New(os.Stdout, name+" ", log.LstdFlags),
		info:     log.New(os.Stdout, name+" ", log.LstdFlags),
		warning:  log.New(os.Stdout, name+" ", log.LstdFlags),
		error:    log.New(os.Stderr, name+" ", log.LstdFlags),
		critical: log.New(os.Stdout, name+" ", log.LstdFlags),
	}
	cl.SetLoggingLevel(level)
	return cl
}

func (cl *ConsoleLogger) Debug(v ...interface{}) {
	if DEBUG >= cl.level {
		cl.debug.Printf("DEBUG: %v", fmt.Sprint(v...))
	}
}

func (cl *ConsoleLogger) Info(v ...interface{}) {
	if INFO >= cl.level {
		cl.info.Printf("INFO: %v", fmt.Sprint(v...))
	}
}

func (cl *ConsoleLogger) Warning(v ...interface{}) {
	if WARNING >= cl.level {
		cl.warning.Printf("WARNING: %v", fmt.Sprint(v...))
	}
}

func (cl *ConsoleLogger) Error(v ...interface{}) {
	if ERROR >= cl.level {
		cl.error.Printf("ERROR: %v", fmt.Sprint(v...))
	}
}

func (cl *ConsoleLogger) Critical(v ...interface{}) {
	cl.critical.Printf("CRITICAL: %v", fmt.Sprint(v...))
}

func (cl *ConsoleLogger) SetLoggingLevel(lvl int) {
	if lvl >= DEBUG && lvl <= CRITICAL {
		cl.level = lvl
		cl.Debug("Setting logging level to ", lvl)
	}
}
