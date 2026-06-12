package asyncapi

import (
	"testing"
)

func TestParseAndResolve(t *testing.T) {
	spec := `
asyncapi: 3.0.0
info:
  title: Test Spec
  version: 1.0.0
channels:
  myChannel:
    address: my-topic
  emptyAddressChannel:
    summary: No address field
operations:
  onMessage:
    action: receive
    channel:
      $ref: '#/channels/myChannel'
    x-integron-steps:
      - name: firstStep
        type: http
        next: ""
  onSend:
    action: send
    channel:
      $ref: '#/channels/myChannel'
`
	doc, err := Parse([]byte(spec))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Test ResolveChannelAddress
	addr, err := doc.ResolveChannelAddress("#/channels/myChannel")
	if err != nil {
		t.Fatalf("ResolveChannelAddress failed: %v", err)
	}
	if addr != "my-topic" {
		t.Errorf("Expected 'my-topic', got '%s'", addr)
	}

	// Test default to channel key if address is empty
	fallbackAddr, err := doc.ResolveChannelAddress("#/channels/emptyAddressChannel")
	if err != nil {
		t.Fatalf("ResolveChannelAddress failed for empty address: %v", err)
	}
	if fallbackAddr != "emptyAddressChannel" {
		t.Errorf("Expected 'emptyAddressChannel', got '%s'", fallbackAddr)
	}

	// Test invalid ref format
	_, err = doc.ResolveChannelAddress("invalid-ref")
	if err == nil {
		t.Errorf("Expected error for invalid reference format")
	}

	// Test Resolve non-existent channel
	_, err = doc.ResolveChannelAddress("#/channels/nonExistent")
	if err == nil {
		t.Errorf("Expected error for non-existent channel reference")
	}

	// Test GetTopicToOperationMap
	topicMap, err := doc.GetTopicToOperationMap()
	if err != nil {
		t.Fatalf("GetTopicToOperationMap failed: %v", err)
	}

	// Verify only "receive" operations are mapped
	op, exists := topicMap["my-topic"]
	if !exists {
		t.Fatalf("Expected topic 'my-topic' in map")
	}
	if op.OperationKey != "onMessage" {
		t.Errorf("Expected OperationKey 'onMessage', got '%s'", op.OperationKey)
	}
	if len(op.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(op.Steps))
	}
}
