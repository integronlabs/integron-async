package helpers

func RemoveNull(input interface{}) interface{} {
	// remove null values from input
	if inputArray, ok := input.([]interface{}); ok {
		var result []interface{}
		for _, val := range inputArray {
			if val != nil {
				result = append(result, RemoveNull(val))
			}
		}
		return result
	}

	if inputMap, ok := input.(map[string]interface{}); ok {
		for key, val := range inputMap {
			if val == nil {
				delete(inputMap, key)
			} else {
				inputMap[key] = RemoveNull(val)
			}
		}
		return inputMap
	}

	return input
}
