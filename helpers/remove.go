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
		result := make(map[string]interface{})
		for key, val := range inputMap {
			if val != nil {
				cleaned := RemoveNull(val)
				if cleaned != nil {
					result[key] = cleaned
				}
			}
		}
		return result
	}

	return input
}
