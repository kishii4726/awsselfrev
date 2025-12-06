package cmd

import (
	"awsselfrev/internal/config"
	"awsselfrev/internal/table"
	"fmt"
	"testing"

	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.ListBucketsOutput), args.Error(1)
}

func (m *MockS3Client) GetBucketEncryption(ctx context.Context, params *s3.GetBucketEncryptionInput, optFns ...func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.GetBucketEncryptionOutput), args.Error(1)
}

func (m *MockS3Client) GetPublicAccessBlock(ctx context.Context, params *s3.GetPublicAccessBlockInput, optFns ...func(*s3.Options)) (*s3.GetPublicAccessBlockOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.GetPublicAccessBlockOutput), args.Error(1)
}

func (m *MockS3Client) GetBucketLifecycleConfiguration(ctx context.Context, params *s3.GetBucketLifecycleConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.GetBucketLifecycleConfigurationOutput), args.Error(1)
}

type MockHTTPStatusError struct {
	StatusCode int
}

func (m MockHTTPStatusError) Error() string       { return fmt.Sprintf("mock status code %d", m.StatusCode) }
func (m MockHTTPStatusError) HTTPStatusCode() int { return m.StatusCode }

func TestCheckBucketConfigurations(t *testing.T) {
	client := new(MockS3Client)
	// テスト用のデータを設定
	buckets := []types.Bucket{{Name: aws.String("test-bucket")}}

	// 404 Error for "Not Found" simulation
	err404 := MockHTTPStatusError{StatusCode: 404}

	client.On("ListBuckets", mock.Anything, mock.Anything, mock.Anything).Return(&s3.ListBucketsOutput{Buckets: buckets}, nil)
	client.On("GetBucketEncryption", mock.Anything, mock.Anything, mock.Anything).Return((*s3.GetBucketEncryptionOutput)(nil), err404)
	client.On("GetPublicAccessBlock", mock.Anything, mock.Anything, mock.Anything).Return((*s3.GetPublicAccessBlockOutput)(nil), err404)
	client.On("GetBucketLifecycleConfiguration", mock.Anything, mock.Anything, mock.Anything).Return((*s3.GetBucketLifecycleConfigurationOutput)(nil), err404)

	// テーブルのセットアップ
	tbl := table.SetTable()
	// ルールのセットアップ
	rules := config.RulesConfig{
		Rules: map[string]config.Rule{
			"s3-encryption":    {Service: "S3", Level: "Alert", Issue: "Bucket encryption is not set"},
			"s3-public-access": {Service: "S3", Level: "Alert", Issue: "Block public access is all off"},
			"s3-lifecycle":     {Service: "S3", Level: "Warning", Issue: "Lifecycle policy is not set"},
		},
	}

	// テスト対象の関数を呼び出し
	checkBucketConfigurations(client, "test-bucket", tbl, rules)

	// テーブルの内容を検証
	assert.Equal(t, 2, tbl.NumLines()) // 2 checks fail (encryption, public access). Lifecycle check skipped for non-log bucket.
}
