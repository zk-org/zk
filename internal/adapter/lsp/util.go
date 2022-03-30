package lsp

import (
	"fmt"
	"net/url"
	"runtime"
	"strings"

	"github.com/mickael-menu/zk/internal/util/errors"
)

func pathToURI(path string) string {
	u := &url.URL{
		Scheme: "file",
		Path:   path,
	}
	return u.String()
}

func uriToPath(uri string) (string, error) {
	s := strings.ReplaceAll(uri, "%5C", "/")
	parsed, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	if parsed.Scheme != "file" {
		return "", errors.New("URI was not a file:// URI")
	}

	if runtime.GOOS == "windows" {
		// In Windows "file:///c:/tmp/foo.md" is parsed to "/c:/tmp/foo.md".
		// Strip the first character to get a valid path.
		if strings.Contains(parsed.Path[1:], ":") {
			// url.Parse() behaves differently with "file:///c:/..." and "file://c:/..."
			return parsed.Path[1:], nil
		} else {
			// if the windows drive is not included in Path it will be in Host
			return parsed.Host + "/" + parsed.Path[1:], nil
		}
	}
	return parsed.Path, nil
}

// jsonBoolean can be unmarshalled from integers or strings.
// Neovim cannot send a boolean easily, so it's useful to support integers too.
type jsonBoolean bool

func (b *jsonBoolean) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "1" || s == "true" {
		*b = true
	} else if s == "0" || s == "false" {
		*b = false
	} else {
		return fmt.Errorf("%s: failed to unmarshal as boolean", s)
	}
	return nil
}
