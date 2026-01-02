package cmd

import (
	"context"
	"fmt"
	"log"

	"awsselfrev/internal/aws/api"
	"awsselfrev/internal/color"
	"awsselfrev/internal/config"
	"awsselfrev/internal/table"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var cloudwatchlogsCmd = &cobra.Command{
	Use:   "cloudwatchlogs",
	Short: "Checks CloudWatch Logs configurations for best practices",
	Long: `This command checks various CloudWatch Logs configurations and best practices such as:
- Log group retention settings`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		rules := config.LoadRules()
		tbl := table.SetTable()
		client := cloudwatchlogs.NewFromConfig(cfg)
		_, _, _ = color.SetLevelColor()

		checkCloudWatchLogsConfigurations(client, tbl, rules)

		table.Render("CloudWatchLogs", tbl)
	},
}

func checkCloudWatchLogsConfigurations(client api.CloudWatchLogsClient, tbl *tablewriter.Table, rules config.RulesConfig) {
	resp, err := client.DescribeLogGroups(context.TODO(), &cloudwatchlogs.DescribeLogGroupsInput{})
	if err != nil {
		log.Fatalf("Failed to describe log groups: %v", err)
	}
	if len(resp.LogGroups) == 0 {
		table.AddRow(tbl, []string{"CloudWatchLogs", "-", "-", "No log groups", "-", "-"})
		return
	}
	for _, logGroup := range resp.LogGroups {
		checkLogGroupRetention(logGroup, tbl, rules)
		checkLogGroupKmsEncryption(logGroup, tbl, rules)
	}
}

func checkLogGroupRetention(logGroup types.LogGroup, tbl *tablewriter.Table, rules config.RulesConfig) {
	rule := rules.Get("cloudwatch-retention")
	if logGroup.RetentionInDays == nil {
		table.AddRow(tbl, []string{rule.Service, "Fail", color.ColorizeLevel(rule.Level), *logGroup.LogGroupName, "Never", rule.Issue})
	} else {
		val := fmt.Sprintf("%d days", *logGroup.RetentionInDays)
		table.AddRow(tbl, []string{rule.Service, "Pass", "-", *logGroup.LogGroupName, val, rule.Issue})
	}
}

func checkLogGroupKmsEncryption(logGroup types.LogGroup, tbl *tablewriter.Table, rules config.RulesConfig) {
	rule := rules.Get("cloudwatch-log-group-encryption")
	if logGroup.KmsKeyId == nil {
		table.AddRow(tbl, []string{rule.Service, "Fail", color.ColorizeLevel(rule.Level), *logGroup.LogGroupName, "Disabled", rule.Issue})
	} else {
		table.AddRow(tbl, []string{rule.Service, "Pass", "-", *logGroup.LogGroupName, "Enabled", rule.Issue})
	}
}

func init() {
	rootCmd.AddCommand(cloudwatchlogsCmd)
}
