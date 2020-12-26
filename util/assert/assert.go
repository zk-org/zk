package assert

import (
	"reflect"
	"testing"
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

func Equal(t *testing.T, actual, expected interface{}) {
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Received %+v (type %v), expected %+v (type %v)", actual, reflect.TypeOf(actual), expected, reflect.TypeOf(expected))
	}
}
