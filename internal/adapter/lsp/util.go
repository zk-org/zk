package lsp

import (
	"net/url"
	"github.com/mickael-menu/zk/internal/util/errors"
)

func pathToURI(path string) string {
	u := &url.URL{
		Scheme:   "file",
		Path:     path,
	}
	return u.String()
}


func uriToPath(uri string) (string, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	if parsed.Scheme != "file" {
		return "", errors.New("URI was not a file:// URI")
	}
	return parsed.Path, nil
}

