package helpers

import (
	"encoding/json"

	"github.com/aymerick/raymond"
	"github.com/mickael-menu/zk/internal/util"
	"github.com/mickael-menu/zk/internal/util/errors"
)

// RegisterJSON registers a {{json}} template helper which serializes its
// parameter to a JSON value.
func RegisterJSON(logger util.Logger) {
	raymond.RegisterHelper("json", func(arg interface{}) string {
		jsonBytes, err := json.Marshal(arg)
		if err != nil {
			logger.Err(errors.Wrapf(err, "%v: not a serializable argument for {{json}}", arg))
			return ""
		}
		return string(jsonBytes)
	})
}
