package steps

import (
	"context"
	"errors"

	"github.com/integronlabs/integron-async/helpers"
)

func runError(ctx context.Context, stepMap map[string]interface{}, stepOutputs map[string]interface{}) (interface{}, string, error) {
	msg, ok := stepMap["message"].(string)
	if !ok || msg == "" {
		// Fallback to the context error if available
		if errStr, exists := stepOutputs["error"].(string); exists && errStr != "" {
			msg = errStr
		} else {
			msg = "error step triggered"
		}
	} else {
		msg = helpers.Replace(msg, stepOutputs)
	}
	return nil, "end", errors.New(msg)
}

func init() {
	RegisterStep("error", runError)
}
