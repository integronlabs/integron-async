package engine

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/integronlabs/integron-async/asyncapi"
	"github.com/integronlabs/integron-async/helpers"
	"github.com/integronlabs/integron-async/steps"
	"github.com/sirupsen/logrus"
)

type Engine struct {
	topicMap map[string]asyncapi.ResolvedOperation
}

// NewEngine creates a new instance of the Integron Async workflow engine
func NewEngine(topicMap map[string]asyncapi.ResolvedOperation) *Engine {
	return &Engine{
		topicMap: topicMap,
	}
}

// ProcessBatch processes a batch of Kafka records and returns any batch processing failures
func (e *Engine) ProcessBatch(ctx context.Context, records []KafkaRecord) BatchResponse {
	var failures []BatchItemFailure

	for _, record := range records {
		logrus.WithContext(ctx).Infof("Processing message: topic=%s, partition=%d, offset=%d", record.Topic, record.Partition, record.Offset)

		err := e.processRecord(ctx, record)
		if err != nil {
			logrus.WithContext(ctx).Errorf("Failed to process message (offset=%d): %v", record.Offset, err)
			failures = append(failures, BatchItemFailure{
				ItemIdentifier: strconv.FormatInt(record.Offset, 10),
			})
		}
	}

	return BatchResponse{
		BatchItemFailures: failures,
	}
}

func (e *Engine) processRecord(ctx context.Context, record KafkaRecord) error {
	// Find matching operation for the record topic
	op, exists := e.topicMap[record.Topic]
	if !exists {
		logrus.WithContext(ctx).Warnf("No AsyncAPI operation configured to receive messages on topic: %s. Skipping message.", record.Topic)
		return nil
	}

	if len(op.Steps) == 0 {
		return fmt.Errorf("no steps defined for operation matching topic '%s'", record.Topic)
	}

	// Base64 decode message value
	decodedVal, err := base64.StdEncoding.DecodeString(record.Value)
	if err != nil {
		return fmt.Errorf("failed to base64 decode record value: %w", err)
	}

	// Try parsing decoded value as JSON
	var payload interface{}
	if err := json.Unmarshal(decodedVal, &payload); err != nil {
		// Fallback to raw string
		payload = string(decodedVal)
	}

	// Base64 decode message key if present
	var decodedKey string
	if record.Key != "" {
		keyBytes, err := base64.StdEncoding.DecodeString(record.Key)
		if err == nil {
			decodedKey = string(keyBytes)
		} else {
			decodedKey = record.Key // fallback to raw string if not base64
		}
	}

	// Construct step outputs context with the initial message
	stepOutputs := make(map[string]interface{})
	stepOutputs["message"] = map[string]interface{}{
		"topic":     record.Topic,
		"partition": record.Partition,
		"offset":    record.Offset,
		"key":       decodedKey,
		"payload":   payload,
	}

	// Execute workflow
	err = RunWorkflow(ctx, op.Steps, stepOutputs, payload)
	if err != nil {
		return fmt.Errorf("workflow execution failed: %w", err)
	}

	return nil
}

// RunWorkflow executes the list of steps sequentially
func RunWorkflow(ctx context.Context, stepsArray []interface{}, stepOutputs map[string]interface{}, initialInput interface{}) error {
	stepsMap, err := helpers.CreateStepsMap(stepsArray)
	if err != nil {
		return err
	}

	firstStepMap, ok := stepsArray[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid first step format")
	}
	currentStepKey, ok := firstStepMap["name"].(string)
	if !ok {
		return fmt.Errorf("missing name in first step definition")
	}

	stepInput := initialInput
	for {
		var next string
		var stepOutput interface{}
		stepOutput, next, err = ProcessStep(ctx, currentStepKey, stepsMap, stepOutputs, stepInput)

		if err != nil {
			// If the failing step is not an error step itself, find and execute the error recovery step
			if currentStepMap, ok := stepsMap[currentStepKey].(map[string]interface{}); !ok || currentStepMap["type"] != "error" {
				var errorStepName string
				for name, stepObj := range stepsMap {
					if sm, ok := stepObj.(map[string]interface{}); ok && sm["type"] == "error" {
						errorStepName = name
						break
					}
				}
				if errorStepName != "" {
					logrus.WithContext(ctx).Warnf("Workflow step '%s' failed. Running error recovery step '%s'.", currentStepKey, errorStepName)
					_, _, _ = ProcessStep(ctx, errorStepName, stepsMap, stepOutputs, err)
				}
			}
			return err
		}

		stepOutputs[currentStepKey] = stepOutput

		if next == "" || next == "end" {
			break
		}
		stepInput = stepOutput
		currentStepKey = next
	}
	return nil
}

// ProcessStep executes a single step from the workflow
func ProcessStep(ctx context.Context, currentStepKey string, stepsMap map[string]interface{}, stepOutputs map[string]interface{}, stepInput interface{}) (interface{}, string, error) {
	step, ok := stepsMap[currentStepKey]
	if !ok {
		return nil, "", fmt.Errorf("invalid step definition for: %s", currentStepKey)
	}
	stepMap, ok := step.(map[string]interface{})
	if !ok {
		return nil, "", fmt.Errorf("invalid step definition for: %s", currentStepKey)
	}

	stepType, ok := stepMap["type"].(string)
	if !ok {
		return nil, "", fmt.Errorf("missing or invalid step type in: %s", currentStepKey)
	}

	handler, exists := steps.GetStepHandler(stepType)
	if !exists {
		return nil, "", fmt.Errorf("unknown step type: %s in step %s", stepType, currentStepKey)
	}

	if stepType == "error" {
		if stepInput != nil {
			if errVal, ok := stepInput.(error); ok {
				stepOutputs["error"] = errVal.Error()
			} else if strVal, ok := stepInput.(string); ok {
				stepOutputs["error"] = strVal
			} else {
				stepOutputs["error"] = fmt.Sprintf("%v", stepInput)
			}
		}
		if handler, exists := steps.GetStepHandler("error"); exists {
			return handler(ctx, stepMap, stepOutputs)
		}
		var err error
		if inputErr, ok := stepInput.(error); ok {
			err = inputErr
		} else if inputStr, ok := stepInput.(string); ok {
			err = errors.New(inputStr)
		} else {
			err = errors.New("error step triggered")
		}
		return nil, "end", err
	}

	stepOutput, next, err := handler(ctx, stepMap, stepOutputs)
	if err != nil {
		return nil, "error", err
	}

	return stepOutput, next, nil
}
