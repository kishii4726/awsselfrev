package cmd

import (
	"awsselfrev/internal/aws/api"
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
		rules := config.LoadRules()
		tbl := table.SetTable()
		client := s3.NewFromConfig(cfg)
		_, _, _ = color.SetLevelColor() // Colors are now handling in table rendering or we just pass strings.

		buckets := s3Internal.ListBuckets(client)
		for _, bucket := range buckets {
			checkBucketConfigurations(client, bucket, tbl, rules)
		}

		table.Render("S3", tbl)
	},
}

func checkBucketConfigurations(client api.S3Client, bucket string, table *tablewriter.Table, rules config.RulesConfig) {
	if !s3Internal.IsBucketEncrypted(client, bucket) {
		rule := rules.Get("s3-encryption")
		table.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), bucket, rule.Issue})
	}
	if !s3Internal.IsBlockPublicAccessEnabled(client, bucket) {
		rule := rules.Get("s3-public-access")
		table.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), bucket, rule.Issue})
	}
	if !s3Internal.IsLifeCycleRuleConfiguredLogBucket(client, bucket) {
		rule := rules.Get("s3-lifecycle")
		table.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), bucket, rule.Issue})
	}
	if !s3Internal.IsObjectLockEnabled(client, bucket) {
		rule := rules.Get("s3-object-lock")
		table.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), bucket, rule.Issue})
	}
}

func init() {
	rootCmd.AddCommand(s3Cmd)
}
