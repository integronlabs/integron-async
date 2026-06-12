package aws

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type Client struct {
	s3Client  *s3.Client
	ssmClient *ssm.Client
}

// NewClient initializes the AWS SDK clients using the default config
func NewClient(ctx context.Context) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS default config: %w", err)
	}

	return &Client{
		s3Client:  s3.NewFromConfig(cfg),
		ssmClient: ssm.NewFromConfig(cfg),
	}, nil
}

// FetchFromS3 fetches the specification document from an S3 bucket
func (c *Client) FetchFromS3(ctx context.Context, bucket, key string) ([]byte, error) {
	output, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 object s3://%s/%s: %w", bucket, key, err)
	}
	defer output.Body.Close()

	data, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read S3 object body: %w", err)
	}

	return data, nil
}

// FetchFromSSM fetches the specification document from SSM Parameter Store
func (c *Client) FetchFromSSM(ctx context.Context, parameterName string) ([]byte, error) {
	withDecryption := true
	output, err := c.ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           &parameterName,
		WithDecryption: &withDecryption,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get SSM parameter '%s': %w", parameterName, err)
	}

	if output.Parameter == nil || output.Parameter.Value == nil {
		return nil, fmt.Errorf("SSM parameter '%s' value is nil", parameterName)
	}

	return []byte(*output.Parameter.Value), nil
}
