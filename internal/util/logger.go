package util

import (
	"log"
	"os"
)

// Logger can be used to report logging messages.
// The native log.Logger type implements this interface.
type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Err(error)
}

// NullLogger is a logger ignoring any input.
var NullLogger = nullLogger{}

type nullLogger struct{}

func (n *nullLogger) Printf(format string, v ...interface{}) {}

func (n *nullLogger) Println(v ...interface{}) {}

func (n *nullLogger) Err(err error) {}

// StdLogger is a logger using the standard logger.
type StdLogger struct {
	*log.Logger
}

func NewStdLogger(prefix string, flags int) StdLogger {
	return StdLogger{log.New(os.Stderr, prefix, flags)}
}

func (l StdLogger) Err(err error) {
	if err != nil {
		l.Printf("warning: %v", err)
	}
}

// ProxyLogger is a logger delegating to an underlying logger.
// Can be used to change the active logger during runtime.
type ProxyLogger struct {
	Logger Logger
}

func NewProxyLogger(logger Logger) *ProxyLogger {
	return &ProxyLogger{logger}
}

func (l *ProxyLogger) Printf(format string, v ...interface{}) {
	l.Logger.Printf(format, v...)
}

func (l *ProxyLogger) Println(v ...interface{}) {
	l.Logger.Println(v...)
}

func (l *ProxyLogger) Err(err error) {
	l.Logger.Err(err)
}
