package cmd

import (
	"awsselfrev/internal/color"
	"awsselfrev/internal/table"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) ListBuckets(input *s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*s3.ListBucketsOutput), args.Error(1)
}

// 他の必要なメソッドもモックする
// 例: HeadBucket, GetBucketEncryption, etc.

func TestCheckBucketConfigurations(t *testing.T) {
	client := new(MockS3Client)
	// テスト用のデータを設定
	buckets := []types.Bucket{{Name: aws.String("test-bucket")}}
	client.On("ListBuckets", &s3.ListBucketsInput{}).Return(&s3.ListBucketsOutput{Buckets: buckets}, nil)

	// テーブルのセットアップ
	tbl := table.SetTable()
	levelWarning, levelAlert := color.SetLevelColor()

	// テスト対象の関数を呼び出し
	checkBucketConfigurations(client, "test-bucket", tbl, levelWarning, levelAlert)

	// テーブルの内容を検証
	assert.Equal(t, 1, len(tbl.Rows())) // テスト用に期待される行数に合わせて調整
}
