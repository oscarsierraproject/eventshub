package logger

const (
	DEBUG = iota
	INFO
	WARNING
	ERROR
	CRITICAL
)

type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warning(v ...interface{})
	Error(v ...interface{})
	Critical(v ...interface{})
}
