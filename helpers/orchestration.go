package helpers

import (
	"fmt"
)

const INVALID_STEP_DEFINITION = "invalid step definition"

func CreateStepsMap(stepsArray []interface{}) (map[string]interface{}, error) {
	steps := make(map[string]interface{})
	for _, v := range stepsArray {
		stepsMap, ok := v.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf(INVALID_STEP_DEFINITION)
		}
		name, ok := stepsMap["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("missing or invalid step name")
		}
		steps[name] = stepsMap
	}
	return steps, nil
}
