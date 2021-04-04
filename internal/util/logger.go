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
