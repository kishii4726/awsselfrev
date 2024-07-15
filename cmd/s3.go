package cmd

import (
	s3Internal "awsselfrev/internal/aws/service/s3"
	"awsselfrev/internal/color"
	"awsselfrev/internal/config"
	"awsselfrev/internal/table"

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
		tbl := table.SetTable()
		client := s3.NewFromConfig(cfg)
		_, levelWarning, levelAlert := color.SetLevelColor()

		buckets := s3Internal.ListBuckets(client)
		for _, bucket := range buckets {
			checkBucketConfigurations(client, bucket, tbl, levelWarning, levelAlert)
		}

		table.Render("S3", tbl)
	},
}

func checkBucketConfigurations(client *s3.Client, bucket string, table *tablewriter.Table, levelWarning, levelAlert string) {
	if !s3Internal.IsBucketEncrypted(client, bucket) {
		table.Append([]string{"S3", levelAlert, "Bucket encryption is not set", bucket})
	}
	if !s3Internal.IsBlockPublicAccessEnabled(client, bucket) {
		table.Append([]string{"S3", levelAlert, "Block public access is all off", bucket})
	}
	if !s3Internal.IsLifeCycleRuleConfiguredLogBucket(client, bucket) {
		table.Append([]string{"S3", levelWarning, "Lifecycle policy is not set", bucket})
	}
}

func init() {
	rootCmd.AddCommand(s3Cmd)
}
