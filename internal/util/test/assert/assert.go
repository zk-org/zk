package assert

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mickael-menu/pretty"
)

func True(t *testing.T, value bool) {
	if !value {
		t.Errorf("Expected to be true")
	}
}

func False(t *testing.T, value bool) {
	if value {
		t.Errorf("Expected to be false")
	}
}

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
	if !(reflect.DeepEqual(actual, expected) || cmp.Equal(actual, expected)) {
		t.Errorf("Received (type %v):\n% #v", reflect.TypeOf(actual), pretty.Formatter(actual))
		t.Errorf("\n---\n")
		t.Errorf("But expected (type %v):\n% #v", reflect.TypeOf(expected), pretty.Formatter(expected))
		t.Errorf("\n---\n")
		t.Errorf("Diff:\n")
		for _, diff := range pretty.Diff(actual, expected) {
			t.Errorf("\t% #v", diff)
		}
	}
}

func NotEqual(t *testing.T, actual, other interface{}) {
	if reflect.DeepEqual(actual, other) || cmp.Equal(actual, other) {
		t.Errorf("Received (type %v):\n% #v", reflect.TypeOf(actual), pretty.Formatter(actual))
		t.Errorf("\n---\n")
		t.Errorf("Expected to be different from (type %v):\n% #v", reflect.TypeOf(other), pretty.Formatter(other))
		t.Errorf("\n---\n")
	}
}

func Regexp(t *testing.T, actual, pattern string) {
	rx := regexp.MustCompile(pattern)
	if !(rx.MatchString(actual)) {
		t.Errorf("Received (type %v):\n% #v", reflect.TypeOf(actual), pretty.Formatter(actual))
		t.Errorf("\n---\n")
		t.Errorf("But expected to match regexp:\n% #v", pretty.Formatter(pattern))
		t.Errorf("\n---\n")
		t.Errorf("Diff:\n")
		for _, diff := range pretty.Diff(actual, pattern) {
			t.Errorf("\t% #v", diff)
		}
	}
}

func toJSON(t *testing.T, obj interface{}) string {
	json, err := json.Marshal(obj)
	// json, err := json.MarshalIndent(obj, "", "  ")
	Nil(t, err)
	return string(json)
}

func Err(t *testing.T, err error, expected string) {
	switch {
	case err == nil:
		t.Errorf("Expected error `%v`, received nil", expected)
	case !strings.Contains(err.Error(), expected):
		t.Errorf("Expected error `%v`, received `%v`", expected, err.Error())
	}
}
