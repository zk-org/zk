package helpers

import (
	"github.com/aymerick/raymond"
)

// RegisterSubstring registers a {{substring}} template helper which extracts a
// substring given a starting index and a length.
//
// {{substring 'A full quote' 2 4}} -> "full"
// {{substring 'A full quote' -5 5}} -> "quote"
//
func RegisterSubstring() {
	raymond.RegisterHelper("substring", func(str string, index int, length int) string {
		if index < 0 {
			index = len(str) + index
		}
		if index >= len(str) {
			return ""
		}
		end := min(index+length, len(str))
		return str[index:end]
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
