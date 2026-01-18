package yaml

import (
	"testing"

	"github.com/zk-org/zk/internal/util/test/assert"
)

// Credit: https://github.com/icza/dyno
func TestConvertToJSONCompatible(t *testing.T) {
	cases := []struct {
		title string // Title of the test case
		v     any    // Input dynamic object
		exp   any    // Expected result
	}{
		{
			title: "nil value",
			v:     nil,
			exp:   nil,
		},
		{
			title: "string value",
			v:     "a",
			exp:   "a",
		},
		{
			title: "map[interfac{}]any value",
			v: map[any]any{
				"s": "s",
				1:   1,
			},
			exp: map[string]any{
				"s": "s",
				"1": 1,
			},
		},
		{
			title: "nested maps and slices",
			v: map[any]any{
				"s": "s",
				1:   1,
				float64(0): []any{
					1,
					"x",
					map[any]any{
						"s": "s",
						2.0: 2,
					},
					map[string]any{
						"s": "s",
						"1": 1,
					},
				},
			},
			exp: map[string]any{
				"s": "s",
				"1": 1,
				"0": []any{
					1,
					"x",
					map[string]any{
						"s": "s",
						"2": 2,
					},
					map[string]any{
						"s": "s",
						"1": 1,
					},
				},
			},
		},
	}

	for _, c := range cases {
		v := ConvertToJSONCompatible(c.v)
		assert.Equal(t, v, c.exp)
	}
}
