package asyncapi

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Parse parses the AsyncAPI YAML document
func Parse(data []byte) (*Document, error) {
	var doc Document
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}
	return &doc, nil
}

// ResolveChannelAddress resolves a channel $ref like "#/channels/channelKey" to the actual channel address
func (d *Document) ResolveChannelAddress(ref string) (string, error) {
	const prefix = "#/channels/"
	if !strings.HasPrefix(ref, prefix) {
		return "", fmt.Errorf("invalid channel reference format: %s (expected prefix %s)", ref, prefix)
	}
	key := strings.TrimPrefix(ref, prefix)
	channel, ok := d.Channels[key]
	if !ok {
		return "", fmt.Errorf("referenced channel key '%s' not found in document channels", key)
	}
	
	// Fall back to channel key if Address is empty
	if channel.Address == "" {
		return key, nil
	}
	return channel.Address, nil
}

// GetTopicToOperationMap resolves the specification operations and maps topics to active receiver operations
func (d *Document) GetTopicToOperationMap() (map[string]ResolvedOperation, error) {
	topicMap := make(map[string]ResolvedOperation)
	for opKey, op := range d.Operations {
		if op.Action != "receive" {
			continue
		}
		topic, err := d.ResolveChannelAddress(op.ChannelRef.Ref)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve channel for operation '%s': %w", opKey, err)
		}
		topicMap[topic] = ResolvedOperation{
			OperationKey: opKey,
			Action:       op.Action,
			Topic:        topic,
			Steps:        op.XSteps,
		}
	}
	return topicMap, nil
}
