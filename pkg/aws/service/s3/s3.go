package service

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func ListBuckets(client *s3.Client) []string {
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

func IsBucketEncrypted(client *s3.Client, bucket string) bool {
	_, err := client.GetBucketEncryption(context.TODO(), &s3.GetBucketEncryptionInput{
		Bucket: aws.String(bucket),
	})
	return handleS3Error(err)
}

func IsBlockPublicAccessEnabled(client *s3.Client, bucket string) bool {
	_, err := client.GetPublicAccessBlock(context.TODO(), &s3.GetPublicAccessBlockInput{
		Bucket: aws.String(bucket),
	})
	return handleS3Error(err)
}

func IsLifeCycleRuleConfiguredLogBucket(client *s3.Client, bucket string) bool {
	if strings.Contains(bucket, "log") {
		_, err := client.GetBucketLifecycleConfiguration(context.TODO(), &s3.GetBucketLifecycleConfigurationInput{
			Bucket: aws.String(bucket),
		})
		return handleS3Error(err)
	}
	return true
}

func handleS3Error(err error) bool {
	if err != nil {
		var re *awshttp.ResponseError
		if errors.As(err, &re) {
			if re.HTTPStatusCode() == 404 {
				return false
			} else if re.HTTPStatusCode() != 301 {
				log.Fatal(err)
			}
		}
	}
	return true
}
