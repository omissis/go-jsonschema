package yamlutils

import "fmt"

// FixMapKeys fixes non-string keys that occur in nested YAML unmarshalling results.
func FixMapKeys(m map[string]interface{}) {
	for k, v := range m {
		m[k] = fixMapKeysIn(v)
	}
}

// Fix non-string keys that occur in nested YAML unmarshalling results.
func fixMapKeysIn(value interface{}) interface{} {
	switch t := value.(type) {
	case []interface{}:
		for i, elem := range t {
			t[i] = fixMapKeysIn(elem)
		}

		return t

	case map[interface{}]interface{}:
		m := map[string]interface{}{}

		for k, v := range t {
			ks, ok := k.(string)
			if !ok {
				ks = fmt.Sprintf("%v", k)
			}

			m[ks] = fixMapKeysIn(v)
		}

		return m

	default:
		return value
	}
}
