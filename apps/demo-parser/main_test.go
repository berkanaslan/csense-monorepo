package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"log"
	"os/exec"
	"testing"
)

const TestFileName = "test_match.dem"

var event = &Event{
	FileName: TestFileName,
	SteamID:  "TEST_STEAM_ID",
}

func setupTestEnvironment() {
	dropTestEnvironment()

	cmd := exec.Command("docker", "run", "-d", "--name=localstack", "-p", "4566:4566", "localstack/localstack")
	err := cmd.Run()

	if err != nil {
		log.Fatal(err)
	}

	// Create a bucket
	cmd = exec.Command("aws", "s3", "mb", "s3://"+BUCKET_NAME, "--endpoint-url", "http://localhost:4566")
	err = cmd.Run()

	if err != nil {
		log.Fatal(err)
	}

	// Upload test file to the bucket
	cmd = exec.Command("aws", "s3", "cp", "assets/"+TestFileName,
		"s3://"+BUCKET_NAME, "--endpoint-url", "http://localhost:4566", "--acl", "public-read", "--region", "us-east-1")
	err = cmd.Run()

	if err != nil {
		log.Fatal(err)
	}
}

func TestDownloadDemo(t *testing.T) {
	// TODO: Implement
}

func TestParseDemo(t *testing.T) {
	setupTestEnvironment()

	cfg, err := getAWSConfig()

	assert.NoError(t, err, "cfg expected to be loaded without error.")
	assert.NotEmpty(t, cfg, "cfg expected to be loaded without error.")

	client := s3.NewFromConfig(cfg)

	assert.NotEmpty(t, client, "client expected to be loaded without error.")

	demo, err := ParseDemo(context.Background(), event, client)

	assert.Empty(t, err, "err expected to be empty.")
	assert.NotEmpty(t, demo, "demo expected to be loaded without error.")

	dropTestEnvironment()
}

func getAWSConfig() (aws.Config, error) {
	return config.LoadDefaultConfig(context.Background(),
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
}

func TestRemoveDemoFromS3(t *testing.T) {
	// TODO: Implement
}

func TestRemoveLocalDemo(t *testing.T) {
	// TODO: Implement
}

func dropTestEnvironment() {
	cmd := exec.Command("docker", "stop", "localstack")
	cmd.Run()

	cmd = exec.Command("docker", "rm", "localstack")
	cmd.Run()
}
