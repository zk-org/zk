package helpers

import (
	"encoding/json"
	"fmt"

	"github.com/aymerick/raymond"
	"github.com/zk-org/zk/internal/util"
)

// RegisterJSON registers a {{json}} template helper which serializes its
// parameter to a JSON value.
func RegisterJSON(logger util.Logger) {
	raymond.RegisterHelper("json", func(arg any) string {
		jsonBytes, err := json.Marshal(arg)
		if err != nil {
			logger.Err(fmt.Errorf("%v: not a serializable argument for {{json}}: %w", arg, err))
			return ""
		}
		return string(jsonBytes)
	})
}
