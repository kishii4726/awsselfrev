package cmd

import (
	"context"
	"log"

	"awsselfrev/internal/aws/api"
	"awsselfrev/internal/color"
	"awsselfrev/internal/config"
	"awsselfrev/internal/table"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var elbCmd = &cobra.Command{
	Use:   "elb",
	Short: "Check ELB configurations for best practices",
	Long: `This command checks various ELB configurations and best practices such as:
- Access logging enabled
- Connection logging enabled
- Deletion protection enabled`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		rules := config.LoadRules()
		tbl := table.SetTable()
		client := elasticloadbalancingv2.NewFromConfig(cfg)
		_, _, _ = color.SetLevelColor()

		checkELBConfigurations(client, tbl, rules)

		table.Render("ELB", tbl)
	},
}

func init() {
	rootCmd.AddCommand(elbCmd)
}

func checkELBConfigurations(client api.ELBv2Client, tbl *tablewriter.Table, rules config.RulesConfig) {
	resp, err := client.DescribeLoadBalancers(context.TODO(), &elasticloadbalancingv2.DescribeLoadBalancersInput{})
	if err != nil {
		log.Fatalf("Failed to describe load balancers: %v", err)
	}

	if len(resp.LoadBalancers) == 0 {
		table.AddRow(tbl, []string{"ELB", "-", "-", "No load balancers", "-", "-"})
		return
	}

	for _, lb := range resp.LoadBalancers {
		if lb.Type != types.LoadBalancerTypeEnumApplication {
			continue
		}

		attrs, err := client.DescribeLoadBalancerAttributes(context.TODO(), &elasticloadbalancingv2.DescribeLoadBalancerAttributesInput{
			LoadBalancerArn: lb.LoadBalancerArn,
		})
		if err != nil {
			log.Fatalf("Failed to describe attributes for ELB %s: %v", *lb.LoadBalancerName, err)
		}

		checkELBAccessLogs(lb, attrs, tbl, rules)
		checkELBConnectionLogs(lb, attrs, tbl, rules)
		checkELBDeletionProtection(lb, attrs, tbl, rules)
	}
}

func checkELBAccessLogs(lb types.LoadBalancer, attrs *elasticloadbalancingv2.DescribeLoadBalancerAttributesOutput, tbl *tablewriter.Table, rules config.RulesConfig) {
	enabled := false
	for _, attr := range attrs.Attributes {
		if *attr.Key == "access_logs.s3.enabled" && *attr.Value == "true" {
			enabled = true
			break
		}
	}
	rule := rules.Get("alb-access-logging")
	if !enabled {
		table.AddRow(tbl, []string{rule.Service, "Fail", color.ColorizeLevel(rule.Level), *lb.LoadBalancerName, "Disabled", rule.Issue})
	} else {
		table.AddRow(tbl, []string{rule.Service, "Pass", "-", *lb.LoadBalancerName, "Enabled", rule.Issue})
	}
}

func checkELBConnectionLogs(lb types.LoadBalancer, attrs *elasticloadbalancingv2.DescribeLoadBalancerAttributesOutput, tbl *tablewriter.Table, rules config.RulesConfig) {
	enabled := false
	for _, attr := range attrs.Attributes {
		if *attr.Key == "connection_logs.s3.enabled" && *attr.Value == "true" {
			enabled = true
			break
		}
	}
	rule := rules.Get("alb-connection-logging")
	if !enabled {
		table.AddRow(tbl, []string{rule.Service, "Fail", color.ColorizeLevel(rule.Level), *lb.LoadBalancerName, "Disabled", rule.Issue})
	} else {
		table.AddRow(tbl, []string{rule.Service, "Pass", "-", *lb.LoadBalancerName, "Enabled", rule.Issue})
	}
}

func checkELBDeletionProtection(lb types.LoadBalancer, attrs *elasticloadbalancingv2.DescribeLoadBalancerAttributesOutput, tbl *tablewriter.Table, rules config.RulesConfig) {
	enabled := false
	for _, attr := range attrs.Attributes {
		if *attr.Key == "deletion_protection.enabled" && *attr.Value == "true" {
			enabled = true
			break
		}
	}
	rule := rules.Get("alb-deletion-protection")
	if !enabled {
		table.AddRow(tbl, []string{rule.Service, "Fail", color.ColorizeLevel(rule.Level), *lb.LoadBalancerName, "Disabled", rule.Issue})
	} else {
		table.AddRow(tbl, []string{rule.Service, "Pass", "-", *lb.LoadBalancerName, "Enabled", rule.Issue})
	}
}
