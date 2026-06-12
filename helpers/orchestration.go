package helpers

import (
	"fmt"
)

const invalidStepDefinition = "invalid step definition"

func CreateStepsMap(stepsArray []interface{}) (map[string]interface{}, error) {
	steps := make(map[string]interface{})
	for _, v := range stepsArray {
		stepsMap, ok := v.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf(invalidStepDefinition)
		}
		name, ok := stepsMap["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("missing or invalid step name")
		}
		if _, exists := steps[name]; exists {
			return nil, fmt.Errorf("duplicate step name: %s", name)
		}
		steps[name] = stepsMap
	}
	return steps, nil
}
