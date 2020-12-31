package assert

import (
	"encoding/json"
	"reflect"
	"testing"
)

func Nil(t *testing.T, value interface{}) {
	if !isNil(value) {
		t.Errorf("Expected `%v` (type %v) to be nil", value, reflect.TypeOf(value))
	}
}

func NotNil(t *testing.T, value interface{}) {
	if isNil(value) {
		t.Errorf("Expected `%v` (type %v) to not be nil", value, reflect.TypeOf(value))
	}
}

func isNil(value interface{}) bool {
	return value == nil ||
		(reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil())
}

func Equal(t *testing.T, actual, expected interface{}) {
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Received (type %v):\n%+v\n---\nBut expected (type %v):\n%+v", reflect.TypeOf(actual), toJSON(t, actual), reflect.TypeOf(expected), toJSON(t, expected))
	}
}

func toJSON(t *testing.T, obj interface{}) string {
	json, err := json.Marshal(obj)
	// json, err := json.MarshalIndent(obj, "", "  ")
	Nil(t, err)
	return string(json)
}
