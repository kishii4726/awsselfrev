/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"log"

	"awsselfrev/pkg/color"
	"awsselfrev/pkg/config"
	"awsselfrev/pkg/table"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/spf13/cobra"
)

// cloudwatchlogsCmd represents the cloudwatchlogs command
var cloudwatchlogsCmd = &cobra.Command{
	Use:   "cloudwatchlogs",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		table := table.SetTable()
		client := cloudwatchlogs.NewFromConfig(cfg)
		// level_info, level_warning, level_alert := color.SetLevelColor()
		_, _, level_alert := color.SetLevelColor()

		resp, err := client.DescribeLogGroups(context.TODO(), &cloudwatchlogs.DescribeLogGroupsInput{})
		if err != nil {
			log.Fatalf("%v", err)
		}
		for _, v := range resp.LogGroups {
			if *&v.RetentionInDays == nil {
				table.Append([]string{"CloudWatchLogs", level_alert, *v.LogGroupName + "の保持期間が設定されていません"})
			}
		}
		table.Render()
	},
}

func init() {
	rootCmd.AddCommand(cloudwatchlogsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cloudwatchlogsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cloudwatchlogsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
