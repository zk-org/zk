package errors

import (
	"fmt"
)

func Wrapper(msg string) func(error) error {
	return func(err error) error {
		return Wrap(err, msg)
	}
}

func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}
