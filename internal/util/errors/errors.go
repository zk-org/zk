package errors

import (
	"fmt"
)

func Wrapperf(format string, args ...any) func(error) error {
	return Wrapper(fmt.Sprintf(format, args...))
}

func Wrapper(msg string) func(error) error {
	return func(err error) error {
		if err == nil {
			return nil
		}
		return fmt.Errorf("%s: %w", msg, err)
	}
}
