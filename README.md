# Integron Async

**Integron Async** is an event-driven, consumer-only workflow execution engine designed for AWS Lambda. It interprets **AsyncAPI v3.0** specifications and runs custom workflows (`x-integron-steps`) in response to Kafka event streams delivered via **AWS EventBridge Pipes**.

---

## Features

- **AsyncAPI v3.0 Specification Support**: Maps channels and operations to step workflows cleanly.
- **Serverless/Lambda Native**: Optimized for AWS Lambda lifecycle and fast cold starts.
- **AWS EventBridge Pipes Integration**: Integrates directly with EventBridge Pipes Kafka batch payloads.
- **Partial Batch Failures**: Reports individual failed offsets using the AWS standard `batchItemFailures` structure, preventing message loss and minimizing duplicate processing.
- **Dynamic Spec Loading**: Automatically loads specification files at startup from AWS SSM Parameter Store or an Amazon S3 Bucket.
- **CLI Mode for Local Testing**: Runs end-to-end workflows locally using simulated event batches.

---

## Workflow Steps

Integron Async ports the core transformations and side-effect engines from the synchronous Integron project:

- `http`: Make synchronous HTTP requests (GET, POST, etc.) and inject response values into the context.
- `transformarray`: Map lists using JSONPath selections.
- `transformobject`: Reshape objects and format the final workflow response.
- `removenull`: Recursively strip out `null` keys or elements from values.
- `error`: Custom step for handling failures or defining error branches.

---

## Specification Example

Workflows are declared under `x-integron-steps` inside the **operation** object in an AsyncAPI v3.0 schema:

```yaml
asyncapi: 3.0.0
info:
  title: Fact Event Listener
  version: 1.0.0

channels:
  dogFactRequests:
    address: dogfact-requests-topic

operations:
  onFactRequest:
    action: receive
    channel:
      $ref: '#/channels/dogFactRequests'
    x-integron-steps:
      - name: fetchFact
        type: http
        url: 'https://dogapi.dog/api/v2/facts?limit=$.message.payload.amount'
        method: GET
        responses:
          '200':
            output:
              response: $.body
            next: "marshalResponse"
      - name: marshalResponse
        type: transformobject
        output:
          data: $.fetchFact.response.data
        next: ""
```

---

## Getting Started

### Prerequisites

- Go 1.22 or higher.

### Local Development (CLI Mode)

You can run the engine locally by passing a simulated EventBridge Pipes JSON batch.

1. **Create a mock input batch** (`test_input.json`):
   ```json
   [
     {
       "eventSource": "SelfManagedKafka",
       "bootstrapServers": "localhost:9092",
       "topic": "dogfact-requests-topic",
       "partition": 0,
       "offset": 1001,
       "value": "eyJhbW91bnQiOiAyfQ=="
     }
   ]
   ```
   *Note: The `value` field represents the base64-encoded string of the payload: `{"amount": 2}`.*

2. **Run the CLI**:
   ```bash
   LOG_LEVEL=debug go run main.go -spec docs/asyncapi.yaml -input test_input.json
   ```

### Running Unit Tests

To run the test suite:
```bash
go test ./...
```

---

## AWS Lambda Deployment

### Build the Binary

Compile the Go binary targeting Linux (standard for Lambda execution):

```bash
GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap main.go
```
*Zip the binary into `deployment.zip` or bundle it in a container image.*

### Environment Variables

Configure the following environment variables on your AWS Lambda function to customize spec loading:

| Variable | Description | Example |
| :--- | :--- | :--- |
| `ASYNCAPI_SPEC_SOURCE` | The location source of the AsyncAPI spec (`S3`, `SSM`, or `LOCAL`). | `S3` |
| `ASYNCAPI_SPEC_S3_BUCKET` | The S3 bucket name containing the spec (required for `S3`). | `my-configs-bucket` |
| `ASYNCAPI_SPEC_S3_KEY` | The path key to the spec in the bucket (required for `S3`). | `asyncapi.yaml` |
| `ASYNCAPI_SPEC_SSM_PARAM` | The SSM parameter name (required for `SSM`). | `/config/asyncapi` |
| `LOG_LEVEL` | Level of logging output (`debug`, `info`, `warn`, `error`). | `info` |

### IAM Permissions

If loading dynamically from S3 or SSM, ensure the Lambda function's execution role has the appropriate permissions:

- **S3 Source**: `s3:GetObject` on the specified bucket key.
- **SSM Source**: `ssm:GetParameter` on the parameter ARN.

### EventBridge Pipes Configuration

Ensure the EventBridge Pipe's target is set to invoke your Lambda function. In the Pipe's target configuration:
1. Turn on **Batching** if you want to group events.
2. If using partial batch responses, check the option to report **Batch Item Failures** in the target settings.
