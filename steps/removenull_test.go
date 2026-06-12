package steps

import (
	"context"
	"testing"
)

func TestRunRemoveNull(t *testing.T) {
	stepMap := map[string]interface{}{
		"name":  "removeNullStep",
		"type":  "removenull",
		"input": "$.message.payload.data",
		"next":  "",
	}

	stepOutputs := map[string]interface{}{
		"message": map[string]interface{}{
			"payload": map[string]interface{}{
				"data": map[string]interface{}{
					"keep":     "value",
					"nilValue": nil,
					"nested": map[string]interface{}{
						"keepNested": "nestedVal",
						"nilNested":  nil,
					},
					"array": []interface{}{
						"elem1",
						nil,
						"elem2",
					},
				},
			},
		},
	}

	output, next, err := runRemoveNull(context.Background(), stepMap, stepOutputs)
	if err != nil {
		t.Fatalf("runRemoveNull failed: %v", err)
	}

	if next != "" {
		t.Errorf("Expected next to be empty, got '%s'", next)
	}

	outputMap, ok := output.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected output to be a map")
	}

	if _, exists := outputMap["nilValue"]; exists {
		t.Errorf("Expected 'nilValue' key to be completely removed from the output map")
	}

	if outputMap["keep"] != "value" {
		t.Errorf("Expected keep to be 'value', got: %v", outputMap["keep"])
	}

	nested := outputMap["nested"].(map[string]interface{})
	if _, exists := nested["nilNested"]; exists {
		t.Errorf("Expected nilNested to be removed")
	}
	if nested["keepNested"] != "nestedVal" {
		t.Errorf("Expected keepNested to be 'nestedVal', got: %v", nested["keepNested"])
	}

	arr := outputMap["array"].([]interface{})
	if len(arr) != 2 {
		t.Fatalf("Expected array length 2, got %d", len(arr))
	}
	if arr[0] != "elem1" || arr[1] != "elem2" {
		t.Errorf("Unexpected array elements: %v", arr)
	}
}
