package steps

import (
	"context"
	"testing"
)

func TestRunTransformObject(t *testing.T) {
	stepMap := map[string]interface{}{
		"name": "objectTransform",
		"type": "transformobject",
		"output": map[string]interface{}{
			"foo": "$.message.payload.value",
			"nested": map[string]interface{}{
				"bar": "$.message.payload.nestedValue",
			},
		},
		"next": "",
	}

	stepOutputs := map[string]interface{}{
		"message": map[string]interface{}{
			"payload": map[string]interface{}{
				"value":       "hello",
				"nestedValue": "world",
			},
		},
	}

	output, next, err := runTransformObject(context.Background(), stepMap, stepOutputs)
	if err != nil {
		t.Fatalf("runTransformObject failed: %v", err)
	}

	if next != "" {
		t.Errorf("Expected next to be empty, got '%s'", next)
	}

	outputMap, ok := output.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected output to be a map")
	}

	if outputMap["foo"] != "hello" {
		t.Errorf("Expected foo to be 'hello', got '%v'", outputMap["foo"])
	}

	nested := outputMap["nested"].(map[string]interface{})
	if nested["bar"] != "world" {
		t.Errorf("Expected nested.bar to be 'world', got '%v'", nested["bar"])
	}
}
