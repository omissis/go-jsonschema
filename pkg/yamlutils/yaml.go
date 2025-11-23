package yamlutils

import "fmt"

// FixMapKeys fixes non-string keys that occur in nested YAML unmarshalling results.
func FixMapKeys(m map[string]any) {
	for k, v := range m {
		m[k] = fixMapKeysIn(v)
	}
}

// Fix non-string keys that occur in nested YAML unmarshalling results.
func fixMapKeysIn(value any) any {
	switch t := value.(type) {
	case []any:
		for i, elem := range t {
			t[i] = fixMapKeysIn(elem)
		}

		return t

	case map[any]any:
		m := map[string]any{}

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
