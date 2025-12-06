package cmd

import (
	"context"
	"log"

	"awsselfrev/internal/aws/api"
	"awsselfrev/internal/color"
	"awsselfrev/internal/config"
	"awsselfrev/internal/table"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
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

		checkLogGroupsRetention(client, tbl, rules)

		table.Render("CloudWatchLogs", tbl)
	},
}

func checkLogGroupsRetention(client api.CloudWatchLogsClient, table *tablewriter.Table, rules config.RulesConfig) {
	resp, err := client.DescribeLogGroups(context.TODO(), &cloudwatchlogs.DescribeLogGroupsInput{})
	if err != nil {
		log.Fatalf("Failed to describe log groups: %v", err)
	}
	for _, logGroup := range resp.LogGroups {
		if logGroup.RetentionInDays == nil {
			rule := rules.Get("cloudwatch-retention")
			table.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), *logGroup.LogGroupName, rule.Issue})
		}
	}
}

func init() {
	rootCmd.AddCommand(cloudwatchlogsCmd)
}
