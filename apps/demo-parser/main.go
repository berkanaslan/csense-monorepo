package main

import (
	"context"
	"github.com/akiver/cs-demo-analyzer/pkg/api"
	"github.com/akiver/cs-demo-analyzer/pkg/api/constants"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	TempDirectory = "/tmp"
)

func main() {
	lambda.Start(func(ctx context.Context, event *Event) (*api.Match, error) {
		cfg, err := config.LoadDefaultConfig(ctx)

		if err != nil {
			return nil, err
		}

		client := s3.NewFromConfig(cfg)
		match, err := ParseDemo(ctx, event, client)

		if err != nil {
			return nil, err
		}

		_ = RemoveDemoFileOnLocal(event)
		_ = RemoveDemoFromS3(ctx, event, client)

		return match, nil
	})
}

// DownloadDemo downloads a demo file from S3 and returns a file object
func DownloadDemo(ctx context.Context, event *Event, s3Client *s3.Client) (*os.File, error) {
	bucketName := BucketName
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	objectOutput, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &event.FileName,
	})

	if err != nil {
		return nil, err
	}

	defer objectOutput.Body.Close()

	localFile, err := os.Create(filepath.Join(TempDirectory, event.FileName))

	if err != nil {
		return nil, err
	}

	defer localFile.Close()

	_, err = io.Copy(localFile, objectOutput.Body)

	if err != nil {
		return nil, err
	}

	return localFile, nil
}

// ParseDemo parses a demo file and returns a Match object's json representation
func ParseDemo(ctx context.Context, event *Event, s3Client *s3.Client) (*api.Match, error) {
	log.Printf("Processing request data for user %s with %s file.", event.SteamID, event.FileName)

	file, err := DownloadDemo(ctx, event, s3Client)

	if err != nil {
		log.Fatalf("Error downloading file: %s", err.Error())
	}

	match, err := api.AnalyzeDemo(file.Name(), api.AnalyzeDemoOptions{
		IncludePositions: false,
		Source:           constants.DemoSourceFaceIt,
	})

	if err != nil {
		log.Fatalf("Error analyzing demo: %s", err.Error())
	}

	return match, nil
}

// RemoveDemoFromS3 deletes a demo file from S3
func RemoveDemoFromS3(ctx context.Context, event *Event, s3Client *s3.Client) error {
	bucketName := BucketName

	_, err := s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &bucketName,
		Key:    &event.FileName,
	})

	return err
}

// RemoveDemoFileOnLocal deletes a demo file from the local filesystem
func RemoveDemoFileOnLocal(event *Event) error {
	return os.Remove(filepath.Join(TempDirectory, event.FileName))
}
