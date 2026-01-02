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

		checkS3Configurations(client, tbl, rules)

		table.Render("S3", tbl)
	},
}

func checkS3Configurations(client api.S3Client, tbl *tablewriter.Table, rules config.RulesConfig) {
	buckets := s3Internal.ListBuckets(client)
	if len(buckets) == 0 {
		table.AddRow(tbl, []string{"S3", "-", "-", "No buckets", "-", "-"})
		return
	}
	for _, bucket := range buckets {
		checkBucketConfigurations(client, bucket, tbl, rules)
	}
}

func checkBucketConfigurations(client api.S3Client, bucket string, tbl *tablewriter.Table, rules config.RulesConfig) {
	ruleEnc := rules.Get("s3-encryption")
	if !s3Internal.IsBucketEncrypted(client, bucket) {
		table.AddRow(tbl, []string{ruleEnc.Service, "Fail", color.ColorizeLevel(ruleEnc.Level), bucket, "Disabled", ruleEnc.Issue})
	} else {
		table.AddRow(tbl, []string{ruleEnc.Service, "Pass", "-", bucket, "Enabled", ruleEnc.Issue})
	}
	rulePub := rules.Get("s3-public-access")
	if !s3Internal.IsBlockPublicAccessEnabled(client, bucket) {
		table.AddRow(tbl, []string{rulePub.Service, "Fail", color.ColorizeLevel(rulePub.Level), bucket, "Disabled", rulePub.Issue})
	} else {
		table.AddRow(tbl, []string{rulePub.Service, "Pass", "-", bucket, "Enabled", rulePub.Issue})
	}
	ruleLife := rules.Get("s3-lifecycle")
	if !s3Internal.IsLifeCycleRuleConfiguredLogBucket(client, bucket) {
		table.AddRow(tbl, []string{ruleLife.Service, "Fail", color.ColorizeLevel(ruleLife.Level), bucket, "Disabled", ruleLife.Issue})
	} else {
		table.AddRow(tbl, []string{ruleLife.Service, "Pass", "-", bucket, "Enabled", ruleLife.Issue})
	}
	ruleLock := rules.Get("s3-object-lock")
	if !s3Internal.IsObjectLockEnabled(client, bucket) {
		table.AddRow(tbl, []string{ruleLock.Service, "Fail", color.ColorizeLevel(ruleLock.Level), bucket, "Disabled", ruleLock.Issue})
	} else {
		table.AddRow(tbl, []string{ruleLock.Service, "Pass", "-", bucket, "Enabled", ruleLock.Issue})
	}
	ruleKms := rules.Get("s3-sse-kms-encryption")
	if !s3Internal.IsBucketEncryptedWithKMS(client, bucket) {
		table.AddRow(tbl, []string{ruleKms.Service, "Fail", color.ColorizeLevel(ruleKms.Level), bucket, "Disabled", ruleKms.Issue})
	} else {
		table.AddRow(tbl, []string{ruleKms.Service, "Pass", "-", bucket, "Enabled", ruleKms.Issue})
	}
	ruleLog := rules.Get("s3-server-access-logging")
	if !s3Internal.IsServerAccessLoggingEnabled(client, bucket) {
		table.AddRow(tbl, []string{ruleLog.Service, "Fail", color.ColorizeLevel(ruleLog.Level), bucket, "Disabled", ruleLog.Issue})
	} else {
		table.AddRow(tbl, []string{ruleLog.Service, "Pass", "-", bucket, "Enabled", ruleLog.Issue})
	}
}

func init() {
	rootCmd.AddCommand(s3Cmd)
}
