package main

import (
	"context"
	"fmt"
	"github.com/akiver/cs-demo-analyzer/pkg/api"
	"github.com/akiver/cs-demo-analyzer/pkg/api/constants"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"os"
)

func main() {
	lambda.Start(ParseDemo)
}

func ParseDemo(ctx context.Context, event *Event) (*api.Match, error) {
	client, err := getS3Client(ctx)

	fmt.Println("Parsing demo: " + event.FileName + " from user: " + event.SteamID)

	if err != nil {
		return nil, err
	}

	bucketName := "csense-demos"

	// Get the object from the bucket
	objectOutput, err := client.GetObject(ctx, &s3.GetObjectInput{
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

func getS3Client(ctx context.Context) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)
	return client, nil
}
