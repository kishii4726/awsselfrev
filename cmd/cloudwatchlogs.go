package cmd

import (
	"context"
	"log"

	"awsselfrev/pkg/color"
	"awsselfrev/pkg/config"
	"awsselfrev/pkg/table"

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
		table := table.SetTable()
		client := cloudwatchlogs.NewFromConfig(cfg)
		_, _, levelAlert := color.SetLevelColor()

		checkLogGroupsRetention(client, table, levelAlert)

		table.Render()
	},
}

func checkLogGroupsRetention(client *cloudwatchlogs.Client, table *tablewriter.Table, levelAlert string) {
	resp, err := client.DescribeLogGroups(context.TODO(), &cloudwatchlogs.DescribeLogGroupsInput{})
	if err != nil {
		log.Fatalf("Failed to describe log groups: %v", err)
	}
	for _, logGroup := range resp.LogGroups {
		if logGroup.RetentionInDays == nil {
			table.Append([]string{"CloudWatchLogs", levelAlert, *logGroup.LogGroupName + "の保持期間が設定されていません"})
		}
	}
}

func init() {
	rootCmd.AddCommand(cloudwatchlogsCmd)
}
