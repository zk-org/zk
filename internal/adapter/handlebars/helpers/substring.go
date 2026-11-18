package helpers

import (
	"github.com/aymerick/raymond"
)

// RegisterSubstring registers a {{substring}} template helper which extracts a
// substring given a starting index and a length.
//
// {{substring 'A full quote' 2 4}} -> "full"
// {{substring 'A full quote' -5 5}} -> "quote"
func RegisterSubstring() {
	raymond.RegisterHelper("substring", func(str string, index int, length int) string {
		runes := []rune(str)
		if index < 0 {
			index = max(len(runes)+index, 0)
		}
		if index >= len(runes) || length <= 0 {
			return ""
		}
		length = min(length, len(runes)-index)
		return string(runes[index : index+length])
	})
}
