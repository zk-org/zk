package yaml

import (
	"testing"

	"github.com/mickael-menu/zk/util/test/assert"
)

// Credit: https://github.com/icza/dyno
func TestConvertToJSONCompatible(t *testing.T) {
	cases := []struct {
		title string      // Title of the test case
		v     interface{} // Input dynamic object
		exp   interface{} // Expected result
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
			title: "map[interfac{}]interface{} value",
			v: map[interface{}]interface{}{
				"s": "s",
				1:   1,
			},
			exp: map[string]interface{}{
				"s": "s",
				"1": 1,
			},
		},
		{
			title: "nested maps and slices",
			v: map[interface{}]interface{}{
				"s": "s",
				1:   1,
				float64(0): []interface{}{
					1,
					"x",
					map[interface{}]interface{}{
						"s": "s",
						2.0: 2,
					},
					map[string]interface{}{
						"s": "s",
						"1": 1,
					},
				},
			},
			exp: map[string]interface{}{
				"s": "s",
				"1": 1,
				"0": []interface{}{
					1,
					"x",
					map[string]interface{}{
						"s": "s",
						"2": 2,
					},
					map[string]interface{}{
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
