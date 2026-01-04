package cmd

import (
	"awsselfrev/internal/aws/api"
	wafv2Internal "awsselfrev/internal/aws/service/wafv2"
	"awsselfrev/internal/color"
	"awsselfrev/internal/config"
	"awsselfrev/internal/table"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	"github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var wafv2Cmd = &cobra.Command{
	Use:   "wafv2",
	Short: "Check AWS WAF v2 configurations",
	Long:  `Check if logging is enabled for WAF v2 Web ACLs (both Regional and CloudFront scopes).`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		rules := config.LoadRules()
		tbl := table.SetTable()

		// Regional client
		client := wafv2.NewFromConfig(cfg)
		// CloudFront client (WAF for CloudFront is always in us-east-1)
		cfCfg := cfg.Copy()
		cfCfg.Region = "us-east-1"
		cfClient := wafv2.NewFromConfig(cfCfg)

		checkWAFV2Configurations(client, cfClient, tbl, rules)

		table.Render("WAF v2", tbl)
	},
}

func checkWAFV2Configurations(client api.WAFV2Client, cfClient api.WAFV2Client, tbl *tablewriter.Table, rules config.RulesConfig) {
	regionalACLs := wafv2Internal.ListWebACLs(client, types.ScopeRegional)
	cfACLs := wafv2Internal.ListWebACLs(cfClient, types.ScopeCloudfront)

	if len(regionalACLs) == 0 && len(cfACLs) == 0 {
		table.AddRow(tbl, []string{"WAFV2", "-", "-", "No Web ACLs", "-", "-"})
		return
	}

	for _, acl := range regionalACLs {
		checkWebACLLogging(client, acl, tbl, rules, "Regional")
	}
	for _, acl := range cfACLs {
		checkWebACLLogging(cfClient, acl, tbl, rules, "CloudFront")
	}
}

func checkWebACLLogging(client api.WAFV2Client, acl wafv2Internal.WebACLInfo, tbl *tablewriter.Table, rules config.RulesConfig, scope string) {
	rule := rules.Get("wafv2-logging-enabled")
	resourceName := fmt.Sprintf("%s (%s)", acl.Name, scope)
	if !wafv2Internal.IsWAFV2LoggingEnabled(client, acl.ARN) {
		table.AddRow(tbl, []string{rule.Service, "Fail", color.ColorizeLevel(rule.Level), resourceName, "Disabled", rule.Issue})
	} else {
		table.AddRow(tbl, []string{rule.Service, "Pass", "-", resourceName, "Enabled", rule.Issue})
	}
}

func init() {
	rootCmd.AddCommand(wafv2Cmd)
}
