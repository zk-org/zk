package lsp

import (
	"fmt"
	"strings"

	"github.com/tliron/kutil/logging"
)

// glspLogger is a Logger wrapping the GLSP one.
// Can be used to change the active logger during runtime.
type glspLogger struct {
	log logging.Logger
}

func newGlspLogger(log logging.Logger) *glspLogger {
	return &glspLogger{log}
}

func (l *glspLogger) Printf(format string, v ...any) {
	l.log.Debugf("zk: "+format, v...)
}

func (l *glspLogger) Println(vs ...any) {
	var message strings.Builder
	message.WriteString("zk: ")
	for i, v := range vs {
		if i > 0 {
			message.WriteString(", ")
		}
		fmt.Fprint(&message, v)
	}
	l.log.Debug(message.String())
}

func (l *glspLogger) Err(err error) {
	if err != nil {
		l.log.Debugf("zk: warning: %v", err)
	}
}
