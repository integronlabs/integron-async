package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/integronlabs/integron-async/asyncapi"
	awsHelper "github.com/integronlabs/integron-async/aws"
	"github.com/integronlabs/integron-async/engine"
	"github.com/integronlabs/integron-async/helpers"
	_ "github.com/integronlabs/integron-async/steps" // registers step handlers
	"github.com/sirupsen/logrus"
)

var (
	workflowEngine *engine.Engine
	initMu         sync.RWMutex
)

func init() {
	helpers.SetupLogging()
}

func loadSpec(ctx context.Context, localSpecPath string) ([]byte, error) {
	specSource := os.Getenv("ASYNCAPI_SPEC_SOURCE")
	switch specSource {
	case "S3":
		bucket := os.Getenv("ASYNCAPI_SPEC_S3_BUCKET")
		key := os.Getenv("ASYNCAPI_SPEC_S3_KEY")
		if bucket == "" || key == "" {
			return nil, errors.New("ASYNCAPI_SPEC_S3_BUCKET and ASYNCAPI_SPEC_S3_KEY must be set when ASYNCAPI_SPEC_SOURCE is S3")
		}
		awsClient, err := awsHelper.NewClient(ctx)
		if err != nil {
			return nil, err
		}
		return awsClient.FetchFromS3(ctx, bucket, key)
	case "SSM":
		param := os.Getenv("ASYNCAPI_SPEC_SSM_PARAM")
		if param == "" {
			return nil, errors.New("ASYNCAPI_SPEC_SSM_PARAM must be set when ASYNCAPI_SPEC_SOURCE is SSM")
		}
		awsClient, err := awsHelper.NewClient(ctx)
		if err != nil {
			return nil, err
		}
		return awsClient.FetchFromSSM(ctx, param)
	case "LOCAL":
		path := os.Getenv("ASYNCAPI_SPEC_LOCAL_PATH")
		if path == "" {
			path = localSpecPath
		}
		return os.ReadFile(path)
	default:
		// Default to local path if not specified
		return os.ReadFile(localSpecPath)
	}
}

func initEngine(ctx context.Context) (*engine.Engine, error) {
	specData, err := loadSpec(ctx, "docs/asyncapi.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load spec: %w", err)
	}

	doc, err := asyncapi.Parse(specData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spec: %w", err)
	}

	topicMap, err := doc.GetTopicToOperationMap()
	if err != nil {
		return nil, fmt.Errorf("failed to map topics: %w", err)
	}

	return engine.NewEngine(topicMap), nil
}

// lambdaHandler is the entrypoint when running as an AWS Lambda function
func lambdaHandler(ctx context.Context, records []engine.KafkaRecord) (engine.BatchResponse, error) {
	initMu.RLock()
	eng := workflowEngine
	initMu.RUnlock()

	if eng == nil {
		initMu.Lock()
		if workflowEngine == nil {
			var err error
			workflowEngine, err = initEngine(ctx)
			if err != nil {
				initMu.Unlock()
				logrus.WithContext(ctx).Errorf("Failed to initialize engine: %v", err)
				return engine.BatchResponse{}, err
			}
		}
		eng = workflowEngine
		initMu.Unlock()
	}

	return eng.ProcessBatch(ctx, records), nil
}

// runCLI processes simulated inputs locally for development and testing
func runCLI(specPath, inputPath string) error {
	ctx := context.Background()
	logrus.Infof("Running in CLI mode. Spec: %s, Input: %s", specPath, inputPath)

	specData, err := loadSpec(ctx, specPath)
	if err != nil {
		return fmt.Errorf("failed to load spec from '%s': %w", specPath, err)
	}

	doc, err := asyncapi.Parse(specData)
	if err != nil {
		return fmt.Errorf("failed to parse spec: %w", err)
	}

	topicMap, err := doc.GetTopicToOperationMap()
	if err != nil {
		return fmt.Errorf("failed to map topics: %w", err)
	}

	cliEngine := engine.NewEngine(topicMap)

	// Read input JSON
	inputData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file '%s': %w", inputPath, err)
	}

	var records []engine.KafkaRecord
	if err := json.Unmarshal(inputData, &records); err != nil {
		return fmt.Errorf("failed to parse input JSON: %w", err)
	}

	response := cliEngine.ProcessBatch(ctx, records)

	// Print result
	respBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal batch response: %w", err)
	}

	fmt.Println("\n--- Batch Processing Result ---")
	fmt.Println(string(respBytes))
	fmt.Println("-------------------------------")

	return nil
}

func main() {
	// Check if running in AWS Lambda
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		lambda.Start(lambdaHandler)
		return
	}

	// CLI Mode
	specPath := flag.String("spec", "docs/asyncapi.yaml", "Path to the AsyncAPI v3 spec file")
	inputPath := flag.String("input", "test_input.json", "Path to the JSON file containing simulated Kafka EventBridge Pipes input records")
	flag.Parse()

	if err := runCLI(*specPath, *inputPath); err != nil {
		logrus.Errorf("CLI Run failed: %v", err)
		os.Exit(1)
	}
}
