package service

import (
	"awsselfrev/internal/aws/api"
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
)

func IsStorageLensEnabled(client api.S3ControlClient, accountID string) bool {
	resp, err := client.ListStorageLensConfigurations(context.TODO(), &s3control.ListStorageLensConfigurationsInput{
		AccountId: aws.String(accountID),
	})
	if err != nil {
		log.Printf("Warning: Failed to list S3 Storage Lens configurations: %v", err)
		return false
	}

	for _, config := range resp.StorageLensConfigurationList {
		if config.IsEnabled {
			return true
		}
	}
	return false
}
