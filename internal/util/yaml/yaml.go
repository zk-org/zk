package yaml

import "fmt"

func ConvertMapToJSONCompatible(m map[string]any) map[string]any {
	res := map[string]any{}

	for k, v := range m {
		res[k] = ConvertToJSONCompatible(v)
	}

	return res
}

// ConvertToJSONCompatible walks the given dynamic object recursively, and
// converts maps with any key type to maps with string key type. This
// function comes handy if you want to marshal a dynamic object into JSON where
// maps with any key type are not allowed.
//
// Recursion is implemented into values of the following types:
//
//	-map[any]any
//	-map[string]any
//	-[]any
//
// When converting map[any]any to map[string]any,
// fmt.Sprint() with default formatting is used to convert the key to a string key.
//
// Credit: https://github.com/icza/dyno
func ConvertToJSONCompatible(v any) any {
	switch x := v.(type) {
	case map[any]any:
		m := map[string]any{}
		for k, v2 := range x {
			switch k2 := k.(type) {
			case string: // Fast check if it's already a string
				m[k2] = ConvertToJSONCompatible(v2)
			default:
				m[fmt.Sprint(k)] = ConvertToJSONCompatible(v2)
			}
		}
		v = m

	case []any:
		for i, v2 := range x {
			x[i] = ConvertToJSONCompatible(v2)
		}

	case map[string]any:
		for k, v2 := range x {
			x[k] = ConvertToJSONCompatible(v2)
		}
	}

	return v
}
