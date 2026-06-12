package engine

// KafkaRecord represents the structure of an incoming Kafka event payload delivered via EventBridge Pipes
type KafkaRecord struct {
	EventSource      string `json:"eventSource"`
	BootstrapServers string `json:"bootstrapServers"`
	Topic            string `json:"topic"`
	Partition        int64  `json:"partition"`
	Offset           int64  `json:"offset"`
	Timestamp        int64  `json:"timestamp"`
	TimestampType    string `json:"timestampType"`
	Value            string `json:"value"` // Base64-encoded
	Key              string `json:"key"`   // Base64-encoded
}

// KafkaItemIdentifier holds the coordinates for a specific Kafka record
type KafkaItemIdentifier struct {
	Topic     string `json:"topic"`
	Partition int64  `json:"partition"`
	Offset    int64  `json:"offset"`
}

// BatchItemFailure represents an individual record offset that failed processing
type BatchItemFailure struct {
	ItemIdentifier interface{} `json:"itemIdentifier"`
}

// BatchResponse is the response format AWS Lambda expects for reporting partial batch failures
type BatchResponse struct {
	BatchItemFailures []BatchItemFailure `json:"batchItemFailures"`
}
