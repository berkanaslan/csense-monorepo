package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"testing"
)

const TestFileName = "test_match.dem"

var event = &Event{
	FileName: TestFileName,
	SteamID:  "TEST_STEAM_ID",
}

func setupTestEnvironment(t *testing.T) {
	cmd := exec.Command("docker", "run", "-d", "--name=localstack", "-p", "4566:4566", "localstack/localstack")
	err := cmd.Run()

	if err != nil {
		t.Fatalf("Failed to start Localstack: %v", err)
	}

	// Create a bucket
	cmd = exec.Command("aws", "s3", "mb", "s3://"+BucketName, "--endpoint-url", "http://localhost:4566")
	err = cmd.Run()

	if err != nil {
		t.Fatalf("Failed to create S3 bucket: %v", err)
	}

	// Upload test file to the bucket
	cmd = exec.Command("aws", "s3", "cp", "assets/"+TestFileName,
		"s3://"+BucketName, "--endpoint-url", "http://localhost:4566", "--acl", "public-read", "--region", "us-east-1")
	err = cmd.Run()

	if err != nil {
		t.Fatalf("Failed to upload test file to S3: %v", err)
	}
}

func dropTestEnvironment(t *testing.T) {
	cmd := exec.Command("docker", "stop", "localstack")
	err := cmd.Run()

	if err != nil {
		t.Fatalf("Failed to stop Localstack: %v", err)
	}

	cmd = exec.Command("docker", "rm", "localstack")
	err = cmd.Run()

	if err != nil {
		t.Fatalf("Failed to remove Localstack container: %v", err)
	}
}

func getS3Client(t *testing.T) *s3.Client {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:               "http://localhost:4566",
					SigningRegion:     "us-east-1",
					SigningName:       "s3",
					PartitionID:       "aws",
					HostnameImmutable: true,
				}, nil
			},
		)))

	assert.NoError(t, err, "cfg expected to be loaded without error.")
	assert.NotEmpty(t, cfg, "cfg expected to be loaded without error.")

	client := s3.NewFromConfig(cfg)

	assert.NotEmpty(t, client, "client expected to be loaded without error.")

	return client
}

func TestDownloadDemo(t *testing.T) {
	setupTestEnvironment(t)

	client := getS3Client(t)

	file, err := DownloadDemo(context.Background(), event, client)

	assert.Empty(t, err, "err expected to be empty.")
	assert.NotEmpty(t, file, "file expected to be loaded without error.")
	assert.Equal(t, TestFileName, file.Name(), "file name expected to be equal to the test file name.")
}

func TestParseDemo(t *testing.T) {
	client := getS3Client(t)

	demo, err := ParseDemo(context.Background(), event, client)

	assert.Empty(t, err, "err expected to be empty.")
	assert.NotEmpty(t, demo, "demo expected to be loaded without error.")
}

func TestRemoveDemoFromS3(t *testing.T) {
	defer dropTestEnvironment(t)

	client := getS3Client(t)
	err := RemoveDemoFromS3(context.Background(), event, client)
	assert.Empty(t, err, "err expected to be empty.")
}

func TestRemoveDemoFileOnLocal(t *testing.T) {
	err := RemoveDemoFileOnLocal(event)
	assert.Empty(t, err, "err expected to be empty.")

	_, err = os.Stat(event.FileName)
	assert.True(t, os.IsNotExist(err), "file expected to be removed.")
}
