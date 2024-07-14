package cmd

import (
	s3Pkg "awsselfrev/pkg/aws/service/s3"
	"awsselfrev/pkg/color"
	"awsselfrev/pkg/config"
	tablePkg "awsselfrev/pkg/table"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var s3Cmd = &cobra.Command{
	Use:   "s3",
	Short: "Check S3 bucket configurations",
	Long: `The "s3" command allows you to check various configurations of your S3 buckets.

It retrieves information about your S3 buckets and checks for encryption, public access block settings,
and lifecycle rules for buckets with 'log' in their names. The results are displayed in a table format.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		table := tablePkg.SetTable()
		client := s3.NewFromConfig(cfg)
		_, levelWarning, levelAlert := color.SetLevelColor()

		buckets := s3Pkg.ListBuckets(client)
		for _, bucket := range buckets {
			checkBucketConfigurations(client, bucket, table, levelWarning, levelAlert)
		}

		table.Render()
	},
}

func checkBucketConfigurations(client *s3.Client, bucket string, table *tablewriter.Table, levelWarning, levelAlert string) {
	if !s3Pkg.IsBucketEncrypted(client, bucket) {
		table.Append([]string{"S3", levelAlert, bucket + "が暗号化されていません"})
	}
	if !s3Pkg.IsBlockPublicAccessEnabled(client, bucket) {
		table.Append([]string{"S3", levelWarning, bucket + "のパブリックブロックアクセスがすべてオフになっています"})
	}
	if !s3Pkg.IsLifeCycleRuleConfiguredLogBucket(client, bucket) {
		table.Append([]string{"S3", levelWarning, bucket + "にライフサイクルルールが設定されていません"})
	}
}

func init() {
	rootCmd.AddCommand(s3Cmd)
}
