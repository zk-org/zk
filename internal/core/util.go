package core

import "strconv"

// lazyStringer implements Stringer and wait for String() to be called the first
// time before computing its value.
type lazyStringer struct {
	value  *string
	render func() string
}

func newLazyStringer(render func() string) *lazyStringer {
	return &lazyStringer{render: render}
}

// String implements Stringer.
func (s *lazyStringer) String() string {
	if s == nil {
		return ""
	}
	if s.value == nil {
		str := s.render()
		s.value = &str
	}
	return *s.value
}

func (s *lazyStringer) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(s.String())), nil
}
