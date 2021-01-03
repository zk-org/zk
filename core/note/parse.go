package note

import (
	"regexp"
	"strings"
)

type Content struct {
	Title string
	Body  string
}

var contentRegex = regexp.MustCompile(`(?m)^#\s+(.+?)\s*$`)

func Parse(content string) Content {
	var res Content

	if match := contentRegex.FindStringSubmatchIndex(content); len(match) >= 4 {
		res.Title = content[match[2]:match[3]]
		res.Body = strings.TrimSpace(content[match[3]:])
	} else {
		res.Body = strings.TrimSpace(content)
	}

	return res
}
