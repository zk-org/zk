package yaml

import "fmt"

func ConvertMapToJSONCompatible(m map[string]interface{}) map[string]interface{} {
	res := map[string]interface{}{}

	for k, v := range m {
		res[k] = ConvertToJSONCompatible(v)
	}

	return res
}

// ConvertToJSONCompatible walks the given dynamic object recursively, and
// converts maps with interface{} key type to maps with string key type. This
// function comes handy if you want to marshal a dynamic object into JSON where
// maps with interface{} key type are not allowed.
//
// Recursion is implemented into values of the following types:
//   -map[interface{}]interface{}
//   -map[string]interface{}
//   -[]interface{}
//
// When converting map[interface{}]interface{} to map[string]interface{},
// fmt.Sprint() with default formatting is used to convert the key to a string key.
//
// Credit: https://github.com/icza/dyno
func ConvertToJSONCompatible(v interface{}) interface{} {
	switch x := v.(type) {
	case map[interface{}]interface{}:
		m := map[string]interface{}{}
		for k, v2 := range x {
			switch k2 := k.(type) {
			case string: // Fast check if it's already a string
				m[k2] = ConvertToJSONCompatible(v2)
			default:
				m[fmt.Sprint(k)] = ConvertToJSONCompatible(v2)
			}
		}
		v = m

	case []interface{}:
		for i, v2 := range x {
			x[i] = ConvertToJSONCompatible(v2)
		}

	case map[string]interface{}:
		for k, v2 := range x {
			x[k] = ConvertToJSONCompatible(v2)
		}
	}

	return v
}
