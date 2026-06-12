package steps

import (
	"context"
	"testing"
)

func TestRunError(t *testing.T) {
	// 1. Test custom message with replacement
	stepMap := map[string]interface{}{
		"name":    "customError",
		"type":    "error",
		"message": "custom error: $.message.payload.reason",
		"next":    "",
	}

	stepOutputs := map[string]interface{}{
		"message": map[string]interface{}{
			"payload": map[string]interface{}{
				"reason": "upstream down",
			},
		},
	}

	_, next, err := runError(context.Background(), stepMap, stepOutputs)
	if err == nil {
		t.Fatalf("Expected runError to return an error, got nil")
	}

	if err.Error() != "custom error: upstream down" {
		t.Errorf("Expected error message 'custom error: upstream down', got '%s'", err.Error())
	}

	if next != "end" {
		t.Errorf("Expected next 'end', got '%s'", next)
	}

	// 2. Test fallback message using stepOutputs["error"]
	stepMapFallback := map[string]interface{}{
		"name": "fallbackError",
		"type": "error",
		"next": "",
	}

	stepOutputsFallback := map[string]interface{}{
		"error": "database connection reset",
	}

	_, _, err = runError(context.Background(), stepMapFallback, stepOutputsFallback)
	if err == nil {
		t.Fatalf("Expected runError to return an error, got nil")
	}

	if err.Error() != "database connection reset" {
		t.Errorf("Expected error message 'database connection reset', got '%s'", err.Error())
	}
}
