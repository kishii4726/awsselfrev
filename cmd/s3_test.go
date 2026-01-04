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
	"github.com/aws/aws-sdk-go-v2/service/s3control"
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

func (m *MockS3Client) GetObjectLockConfiguration(ctx context.Context, params *s3.GetObjectLockConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetObjectLockConfigurationOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.GetObjectLockConfigurationOutput), args.Error(1)
}

func (m *MockS3Client) GetBucketLogging(ctx context.Context, params *s3.GetBucketLoggingInput, optFns ...func(*s3.Options)) (*s3.GetBucketLoggingOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.GetBucketLoggingOutput), args.Error(1)
}

type MockS3ControlClient struct {
	mock.Mock
}

func (m *MockS3ControlClient) ListStorageLensConfigurations(ctx context.Context, params *s3control.ListStorageLensConfigurationsInput, optFns ...func(*s3control.Options)) (*s3control.ListStorageLensConfigurationsOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3control.ListStorageLensConfigurationsOutput), args.Error(1)
}

type MockHTTPStatusError struct {
	StatusCode int
}

func (m MockHTTPStatusError) Error() string       { return fmt.Sprintf("mock status code %d", m.StatusCode) }
func (m MockHTTPStatusError) HTTPStatusCode() int { return m.StatusCode }

func TestCheckBucketConfigurations(t *testing.T) {
	client := new(MockS3Client)
	controlClient := new(MockS3ControlClient)
	// テスト用のデータを設定 (Use log bucket to trigger all checks)
	buckets := []types.Bucket{
		{Name: aws.String("test-log-bucket")},
		{Name: aws.String("test-bucket")},
	}

	// 404 Error for "Not Found" simulation
	err404 := MockHTTPStatusError{StatusCode: 404}

	client.On("ListBuckets", mock.Anything, mock.Anything, mock.Anything).Return(&s3.ListBucketsOutput{Buckets: buckets}, nil)
	client.On("GetBucketEncryption", mock.Anything, mock.Anything, mock.Anything).Return((*s3.GetBucketEncryptionOutput)(nil), err404)
	client.On("GetPublicAccessBlock", mock.Anything, mock.Anything, mock.Anything).Return((*s3.GetPublicAccessBlockOutput)(nil), err404)
	client.On("GetBucketLifecycleConfiguration", mock.Anything, mock.Anything, mock.Anything).Return((*s3.GetBucketLifecycleConfigurationOutput)(nil), err404)
	client.On("GetObjectLockConfiguration", mock.Anything, mock.Anything, mock.Anything).Return((*s3.GetObjectLockConfigurationOutput)(nil), err404)
	client.On("GetBucketLogging", mock.Anything, mock.Anything, mock.Anything).Return((*s3.GetBucketLoggingOutput)(nil), err404)
	controlClient.On("ListStorageLensConfigurations", mock.Anything, mock.Anything, mock.Anything).Return(&s3control.ListStorageLensConfigurationsOutput{}, nil)

	// テーブルのセットアップ
	tbl := table.SetTable()
	// ルールのセットアップ
	rules := config.RulesConfig{
		Rules: map[string]config.Rule{
			"s3-encryption":            {Service: "S3", Level: "Alert", Issue: "Bucket encryption is not set"},
			"s3-public-access":         {Service: "S3", Level: "Alert", Issue: "Block public access is all off"},
			"s3-lifecycle":             {Service: "S3", Level: "Warning", Issue: "Lifecycle policy is not set"},
			"s3-object-lock":           {Service: "S3", Level: "Warning", Issue: "Object Lock is not enabled"},
			"s3-sse-kms-encryption":    {Service: "S3", Level: "Warning", Issue: "SSE-KMS encryption is not set"},
			"s3-server-access-logging": {Service: "S3", Level: "Warning", Issue: "Server access logging is not enabled"},
			"s3-storage-lens-enabled":  {Service: "S3", Level: "Warning", Issue: "S3 Storage Lens is not enabled"},
		},
	}

	// テスト対象の関数を呼び出し
	checkS3Configurations(client, controlClient, tbl, rules)

	// テーブルの内容を検証
	// Storage Lens: 1 check
	// test-log-bucket: Encryption, Public, Lifecycle, ObjectLock, SSE-KMS, AccessLogs (6 checks)
	// test-bucket: Encryption, Public, Lifecycle, ObjectLock, SSE-KMS, AccessLogs (6 checks)
	// Total rows = 1 + 6 + 6 = 13
	assert.Equal(t, 13, tbl.NumLines())
}
