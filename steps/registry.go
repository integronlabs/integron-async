package steps

import (
	"context"
	"sync"
)

// StepHandler defines the signature for any workflow step runner
type StepHandler func(ctx context.Context, stepMap map[string]interface{}, stepOutputs map[string]interface{}) (interface{}, string, error)

var (
	stepRegistry  = make(map[string]StepHandler)
	registryMutex sync.RWMutex
)

// RegisterStep registers a step handler for a specific step type
func RegisterStep(stepType string, handler StepHandler) {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	stepRegistry[stepType] = handler
}

// GetStepHandler retrieves a step handler by its type
func GetStepHandler(stepType string) (StepHandler, bool) {
	registryMutex.RLock()
	defer registryMutex.RUnlock()
	handler, exists := stepRegistry[stepType]
	return handler, exists
}
