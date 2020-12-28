package opt

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

// OrDefault returns the optional String value or the given default string if it is null.
func (s String) OrDefault(def string) string {
	if s.value == nil {
		return def
	} else {
		return *s.value
	}
}

// Unwrap returns the optional String value or an empty String if none is set.
func (s String) Unwrap() string {
	return s.OrDefault("")
}

func (s String) String() string {
	return s.OrDefault("")
}
