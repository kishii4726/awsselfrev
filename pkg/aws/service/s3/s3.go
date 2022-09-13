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
	var l []string
	resp, err := client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range resp.Buckets {
		l = append(l, *v.Name)
	}
	return l
}

func IsBucketEncrypted(client *s3.Client, bucket string) bool {
	_, err := client.GetBucketEncryption(context.TODO(), &s3.GetBucketEncryptionInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		var re *awshttp.ResponseError
		if errors.As(err, &re) {
			if re.HTTPStatusCode() == 404 {
				return false
			} else if re.HTTPStatusCode() == 301 {
			} else {
				log.Fatal(err)
			}
		}
	}
	return true
}

func IsBlockPublicAccessEnabled(client *s3.Client, bucket string) bool {
	_, err := client.GetPublicAccessBlock(context.TODO(), &s3.GetPublicAccessBlockInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		var re *awshttp.ResponseError
		if errors.As(err, &re) {
			if re.HTTPStatusCode() == 404 {
				return false
			} else if re.HTTPStatusCode() == 301 {
			} else {
				log.Fatal(err)
			}
		}
	}
	return true
}

func IsLifeCycleRuleConfiguredLogBucket(client *s3.Client, bucket string) bool {
	if strings.Contains(bucket, "log") {
		_, err := client.GetBucketLifecycleConfiguration(context.TODO(), &s3.GetBucketLifecycleConfigurationInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			var re *awshttp.ResponseError
			if errors.As(err, &re) {
				if re.HTTPStatusCode() == 404 {
					return false
				} else if re.HTTPStatusCode() == 301 {
				} else {
					log.Fatal(err)
				}
			}
		}
	}
	return true
}
