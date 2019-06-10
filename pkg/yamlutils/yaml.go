package yamlutils

// Fix non-string keys that occur in nested YAML unmarshaling results.
func FixMapKeys(m map[string]interface{}) {
	for k, v := range m {
		m[k] = fixMapKeysIn(v)
	}
}

// Fix non-string keys that occur in nested YAML unmarshaling results.
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
			m[k.(string)] = fixMapKeysIn(v)
		}
		return m
	default:
		return value
	}
}
