package steps

import (
	"context"
	"testing"
)

func TestRunTransformArray(t *testing.T) {
	stepMap := map[string]interface{}{
		"name":  "arrayTransform",
		"type":  "transformarray",
		"input": "$.message.payload.items",
		"output": map[string]interface{}{
			"id":   "$.id",
			"name": "$.name",
		},
		"next": "",
	}

	stepOutputs := map[string]interface{}{
		"message": map[string]interface{}{
			"payload": map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"id": "1", "name": "foo", "extra": "ignored"},
					map[string]interface{}{"id": "2", "name": "bar", "extra": "ignored"},
				},
			},
		},
	}

	output, next, err := runTransformArray(context.Background(), stepMap, stepOutputs)
	if err != nil {
		t.Fatalf("runTransformArray failed: %v", err)
	}

	if next != "" {
		t.Errorf("Expected next to be empty, got '%s'", next)
	}

	outputSlice, ok := output.([]interface{})
	if !ok {
		t.Fatalf("Expected output to be a slice")
	}

	if len(outputSlice) != 2 {
		t.Fatalf("Expected slice length 2, got %d", len(outputSlice))
	}

	item0 := outputSlice[0].(map[string]interface{})
	if item0["id"] != "1" || item0["name"] != "foo" {
		t.Errorf("Unexpected values in item0: %v", item0)
	}
	if _, exists := item0["extra"]; exists {
		t.Errorf("Expected 'extra' key to be completely omitted from transformed item, but it exists")
	}
}
