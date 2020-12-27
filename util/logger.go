package util

// Logger can be used to report logging messages.
// The native log.Logger type implements this interface.
type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

// NullLogger is a logger ignoring any input.
var NullLogger = nullLogger{}

type nullLogger struct{}

func (n *nullLogger) Printf(format string, v ...interface{}) {}

func (n *nullLogger) Println(v ...interface{}) {}
