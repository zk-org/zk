package opt

// String holds an optional string value.
type String struct {
	value *string
}

// NullString repreents an empty optional String.
var NullString = String{nil}

// NewString creates a new optional String with the given value.
func NewString(value string) String {
	return String{&value}
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
