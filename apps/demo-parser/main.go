package main

import (
	"context"
	"github.com/akiver/cs-demo-analyzer/pkg/api"
	"github.com/akiver/cs-demo-analyzer/pkg/api/constants"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"os"
)

func main() {
	lambda.Start(func(ctx context.Context, event *Event) (*api.Match, error) {
		cfg, err := config.LoadDefaultConfig(ctx)

		if err != nil {
			return nil, err
		}

		client := s3.NewFromConfig(cfg)
		return ParseDemo(ctx, event, client)
	})
}

// ParseDemo parses a demo file and returns a Match object's json representation
func ParseDemo(ctx context.Context, event *Event, s3Client *s3.Client) (*api.Match, error) {
	bucketName := "csense-demos"

	objectOutput, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &event.FileName,
	})

	if err != nil {
		return nil, err
	}

	defer objectOutput.Body.Close()

	localFile, err := os.Create(event.FileName)

	if err != nil {
		return nil, err
	}

	defer localFile.Close()

	_, err = io.Copy(localFile, objectOutput.Body)

	if err != nil {
		return nil, err
	}

	match, err := api.AnalyzeDemo(event.FileName, api.AnalyzeDemoOptions{
		IncludePositions: false,
		Source:           constants.DemoSourceValve,
	})

	if err != nil {
		return nil, err
	}

	return match, nil
}
