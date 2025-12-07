package service

import (
	"awsselfrev/internal/aws/api"
	"context"
	"errors"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func ListBuckets(client api.S3Client) []string {
	var buckets []string
	resp, err := client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		log.Fatal(err)
	}
	for _, bucket := range resp.Buckets {
		buckets = append(buckets, *bucket.Name)
	}
	return buckets
}

func IsBucketEncrypted(client api.S3Client, bucket string) bool {
	_, err := client.GetBucketEncryption(context.TODO(), &s3.GetBucketEncryptionInput{
		Bucket: aws.String(bucket),
	})
	return handleS3Error(err)
}

func IsBlockPublicAccessEnabled(client api.S3Client, bucket string) bool {
	_, err := client.GetPublicAccessBlock(context.TODO(), &s3.GetPublicAccessBlockInput{
		Bucket: aws.String(bucket),
	})
	return handleS3Error(err)
}

func IsLifeCycleRuleConfiguredLogBucket(client api.S3Client, bucket string) bool {
	if strings.Contains(bucket, "log") {
		_, err := client.GetBucketLifecycleConfiguration(context.TODO(), &s3.GetBucketLifecycleConfigurationInput{
			Bucket: aws.String(bucket),
		})
		return handleS3Error(err)
	}
	return true
}

func IsObjectLockEnabled(client api.S3Client, bucket string) bool {
	if strings.Contains(bucket, "log") {
		resp, err := client.GetObjectLockConfiguration(context.TODO(), &s3.GetObjectLockConfigurationInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			return handleS3Error(err)
		}
		return resp.ObjectLockConfiguration != nil && resp.ObjectLockConfiguration.ObjectLockEnabled == types.ObjectLockEnabledEnabled
	}
	return true
}

type HTTPStatusError interface {
	HTTPStatusCode() int
}

func handleS3Error(err error) bool {
	if err != nil {
		var se HTTPStatusError
		if errors.As(err, &se) {
			if se.HTTPStatusCode() == 404 {
				return false
			} else if se.HTTPStatusCode() != 301 {
				log.Fatal(err)
			}
		}
	}
	return true
}
