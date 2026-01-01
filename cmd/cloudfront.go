package cmd

import (
	"context"
	"log"

	"awsselfrev/internal/aws/api"
	"awsselfrev/internal/color"
	"awsselfrev/internal/config"
	"awsselfrev/internal/table"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var cloudfrontCmd = &cobra.Command{
	Use:   "cloudfront",
	Short: "Check CloudFront configurations for best practices",
	Long: `This command checks various CloudFront configurations and best practices such as:
- Logging enabled (Standard or Real-time)`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		rules := config.LoadRules()
		tbl := table.SetTable()
		client := cloudfront.NewFromConfig(cfg)
		_, _, _ = color.SetLevelColor()

		checkCloudFrontDistributions(client, tbl, rules)

		table.Render("CloudFront", tbl)
	},
}

func init() {
	rootCmd.AddCommand(cloudfrontCmd)
}

func checkCloudFrontDistributions(client api.CloudFrontClient, table *tablewriter.Table, rules config.RulesConfig) {
	resp, err := client.ListDistributions(context.TODO(), &cloudfront.ListDistributionsInput{})
	if err != nil {
		log.Fatalf("Failed to list CloudFront distributions: %v", err)
	}

	if resp.DistributionList != nil {
		for _, distSummary := range resp.DistributionList.Items {
			checkLoggingEnabled(client, distSummary.Id, table, rules)
		}
	}
}

// checkLoggingEnabled checks if either Standard Logging or Real-time Logging is enabled using GetDistributionConfig
func checkLoggingEnabled(client api.CloudFrontClient, distID *string, table *tablewriter.Table, rules config.RulesConfig) {
	if distID == nil {
		return
	}

	configResp, err := client.GetDistributionConfig(context.TODO(), &cloudfront.GetDistributionConfigInput{
		Id: distID,
	})
	if err != nil {
		log.Printf("Failed to get distribution config for %s: %v", *distID, err)
		return
	}

	distConfig := configResp.DistributionConfig
	if distConfig == nil {
		return
	}

	// 1. Standard Logging
	standardLogging := false
	if distConfig.Logging != nil && distConfig.Logging.Enabled != nil && *distConfig.Logging.Enabled {
		standardLogging = true
	}

	// 2. Real-time Logging
	// Check Default Cache Behavior
	realtimeLogging := false
	if distConfig.DefaultCacheBehavior != nil && distConfig.DefaultCacheBehavior.RealtimeLogConfigArn != nil && *distConfig.DefaultCacheBehavior.RealtimeLogConfigArn != "" {
		realtimeLogging = true
	}

	// Check other Cache Behaviors if default doesn't have it
	if !realtimeLogging && distConfig.CacheBehaviors != nil {
		for _, behavior := range distConfig.CacheBehaviors.Items {
			if behavior.RealtimeLogConfigArn != nil && *behavior.RealtimeLogConfigArn != "" {
				realtimeLogging = true
				break
			}
		}
	}

	rule := rules.Get("cloudfront-logging-enabled")
	if !standardLogging && !realtimeLogging {
		table.Append([]string{rule.Service, "NG", color.ColorizeLevel(rule.Level), *distID, rule.Issue})
	} else {
		table.Append([]string{rule.Service, "Pass", "-", *distID, rule.Issue})
	}
}
