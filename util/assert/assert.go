package assert

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Nil(t *testing.T, value interface{}) {
	if !isNil(value) {
		t.Errorf("Expected %v (type %v) to be nil", value, reflect.TypeOf(value))
	}
}

func NotNil(t *testing.T, value interface{}) {
	if isNil(value) {
		t.Errorf("Expected %v (type %v) to not be nil", value, reflect.TypeOf(value))
	}
}

func isNil(value interface{}) bool {
	return value == nil ||
		(reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil())
}

func Equal(t *testing.T, value interface{}, expected interface{}) {
	if !cmp.Equal(value, expected) {
		t.Errorf("Received %+v (type %v), expected %+v (type %v)", value, reflect.TypeOf(value), expected, reflect.TypeOf(expected))
	}
}
