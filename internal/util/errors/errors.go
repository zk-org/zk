package errors

import (
	"errors"
	"fmt"
)

func Wrapperf(format string, args ...interface{}) func(error) error {
	return Wrapper(fmt.Sprintf(format, args...))
}

func Wrapper(msg string) func(error) error {
	return func(err error) error {
		return Wrap(err, msg)
	}
}

func Wrapf(err error, format string, args ...interface{}) error {
	return Wrap(err, fmt.Sprintf(format, args...))
}

func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

func New(text string) error {
	return errors.New(text)
}

func As(err error, target interface{}) bool {
	return errors.As(err, target)
}
