package engine

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/integronlabs/integron-async/asyncapi"
	_ "github.com/integronlabs/integron-async/steps" // registers step handlers
)

func TestEngineProcessBatch(t *testing.T) {
	// Define a simple workflow with a transformobject step
	steps := []interface{}{
		map[string]interface{}{
			"name": "transformStep",
			"type": "transformobject",
			"output": map[string]interface{}{
				"echo": "$.message.payload.value",
			},
			"next": "",
		},
	}

	topicMap := map[string]asyncapi.ResolvedOperation{
		"test-topic": {
			OperationKey: "op1",
			Action:       "receive",
			Topic:        "test-topic",
			Steps:        steps,
		},
	}

	eng := NewEngine(topicMap)

	// Valid payload: {"value": "hello"}
	valBase64 := base64.StdEncoding.EncodeToString([]byte(`{"value": "hello"}`))

	records := []KafkaRecord{
		{
			Topic:     "test-topic",
			Partition: 0,
			Offset:    100,
			Value:     valBase64,
		},
		{
			Topic:     "test-topic",
			Partition: 0,
			Offset:    101,
			Value:     "invalid-base64-!!!", // Base64 decode will fail -> should report failure
		},
		{
			Topic:     "unknown-topic",
			Partition: 0,
			Offset:    102,
			Value:     valBase64, // Unconfigured topic -> should skip/warn but NOT fail the batch
		},
	}

	resp := eng.ProcessBatch(context.Background(), records)

	// We expect offset 101 to fail, and offsets 100 and 102 to succeed (102 is skipped)
	if len(resp.BatchItemFailures) != 1 {
		t.Fatalf("Expected exactly 1 batch item failure, got %d", len(resp.BatchItemFailures))
	}

	if resp.BatchItemFailures[0].ItemIdentifier != "101" {
		t.Errorf("Expected failed item identifier to be '101', got '%s'", resp.BatchItemFailures[0].ItemIdentifier)
	}
}
