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
	ruleEnc := rules.Get("s3-encryption")
	if !s3Internal.IsBucketEncrypted(client, bucket) {
		table.Append([]string{ruleEnc.Service, "NG", color.ColorizeLevel(ruleEnc.Level), bucket, ruleEnc.Issue})
	} else {
		table.Append([]string{ruleEnc.Service, "Pass", "-", bucket, ruleEnc.Issue})
	}
	rulePub := rules.Get("s3-public-access")
	if !s3Internal.IsBlockPublicAccessEnabled(client, bucket) {
		table.Append([]string{rulePub.Service, "NG", color.ColorizeLevel(rulePub.Level), bucket, rulePub.Issue})
	} else {
		table.Append([]string{rulePub.Service, "Pass", "-", bucket, rulePub.Issue})
	}
	ruleLife := rules.Get("s3-lifecycle")
	if !s3Internal.IsLifeCycleRuleConfiguredLogBucket(client, bucket) {
		table.Append([]string{ruleLife.Service, "NG", color.ColorizeLevel(ruleLife.Level), bucket, ruleLife.Issue})
	} else {
		table.Append([]string{ruleLife.Service, "Pass", "-", bucket, ruleLife.Issue})
	}
	ruleLock := rules.Get("s3-object-lock")
	if !s3Internal.IsObjectLockEnabled(client, bucket) {
		table.Append([]string{ruleLock.Service, "NG", color.ColorizeLevel(ruleLock.Level), bucket, ruleLock.Issue})
	} else {
		table.Append([]string{ruleLock.Service, "Pass", "-", bucket, ruleLock.Issue})
	}
	ruleKms := rules.Get("s3-sse-kms-encryption")
	if !s3Internal.IsBucketEncryptedWithKMS(client, bucket) {
		table.Append([]string{ruleKms.Service, "NG", color.ColorizeLevel(ruleKms.Level), bucket, ruleKms.Issue})
	} else {
		table.Append([]string{ruleKms.Service, "Pass", "-", bucket, ruleKms.Issue})
	}
	ruleLog := rules.Get("s3-server-access-logging")
	if !s3Internal.IsServerAccessLoggingEnabled(client, bucket) {
		table.Append([]string{ruleLog.Service, "NG", color.ColorizeLevel(ruleLog.Level), bucket, ruleLog.Issue})
	} else {
		table.Append([]string{ruleLog.Service, "Pass", "-", bucket, ruleLog.Issue})
	}
}

func init() {
	rootCmd.AddCommand(s3Cmd)
}
