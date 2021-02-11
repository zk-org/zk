package opt

import "fmt"

// String holds an optional string value.
type String struct {
	value *string
}

// NullString represents an empty optional String.
var NullString = String{nil}

// NewString creates a new optional String with the given value.
func NewString(value string) String {
	return String{&value}
}

// NewString creates a new optional String with the given pointer.
// When nil, the String is considered null, but an empty String is valid.
func NewStringWithPtr(value *string) String {
	return String{value}
}

// NewNotEmptyString creates a new optional String with the given value or
// returns NullString if the value is an empty string.
func NewNotEmptyString(value string) String {
	if value == "" {
		return NullString
	} else {
		return NewString(value)
	}
}

// IsNull returns whether the optional String has no value.
func (s String) IsNull() bool {
	return s.value == nil
}

// IsEmpty returns whether the optional String has an empty string for value.
func (s String) IsEmpty() bool {
	return !s.IsNull() && *s.value == ""
}

// Or returns the receiver if it is not null, otherwise the given optional
// String.
func (s String) Or(other String) String {
	if s.IsNull() {
		return other
	} else {
		return s
	}
}

// OrDefault returns the optional String value or the given default string if
// it is null.
func (s String) OrDefault(def string) string {
	if s.IsNull() {
		return def
	} else {
		return *s.value
	}
}

// Unwrap returns the optional String value or an empty String if none is set.
func (s String) Unwrap() string {
	return s.OrDefault("")
}

func (s String) Equal(other String) bool {
	return s.value == other.value ||
		(s.value != nil && other.value != nil && *s.value == *other.value)
}

func (s String) String() string {
	return s.OrDefault("")
}

func (s String) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%v"`, s)), nil
}
