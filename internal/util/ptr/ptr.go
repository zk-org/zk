// Package ptr provides helpers to work with pointers to primitive values.
package ptr

// String returns a pointer to the given string.
func String(v string) *string {
	return &v
}

// NotEmptyString returns a pointer to the given string, or nil if it is empty.
func NotEmptyString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

// Bool returns a pointer to the given boolean.
func Bool(v bool) *bool {
	return &v
}

// Value returns the value pointed to by p, or the zero value if p is nil.
func Value(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// OrString returns the value pointed to by p, or def if p is nil.
func OrString(p *string, def string) string {
	if p == nil {
		return def
	}
	return *p
}

// First returns the first non-nil pointer among the given candidates.
func First[T any](p *T, rest ...*T) *T {
	if p != nil {
		return p
	}
	for _, r := range rest {
		if r != nil {
			return r
		}
	}
	return nil
}
