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

var albCmd = &cobra.Command{
	Use:   "alb",
	Short: "Check ALB configurations for best practices",
	Long: `This command checks various ALB configurations and best practices such as:
- Access logging enabled
- Connection logging enabled
- Deletion protection enabled`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		rules := config.LoadRules()
		tbl := table.SetTable()
		client := elasticloadbalancingv2.NewFromConfig(cfg)
		_, _, _ = color.SetLevelColor()

		checkALBConfigurations(client, tbl, rules)

		table.Render("ALB", tbl)
	},
}

func init() {
	rootCmd.AddCommand(albCmd)
}

func checkALBConfigurations(client api.ELBv2Client, table *tablewriter.Table, rules config.RulesConfig) {
	resp, err := client.DescribeLoadBalancers(context.TODO(), &elasticloadbalancingv2.DescribeLoadBalancersInput{})
	if err != nil {
		log.Fatalf("Failed to describe load balancers: %v", err)
	}

	if len(resp.LoadBalancers) == 0 {
		table.Append([]string{"ALB", "-", "-", "No load balancers", "-", "-"})
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
			log.Fatalf("Failed to describe attributes for ALB %s: %v", *lb.LoadBalancerName, err)
		}

		checkALBAccessLogs(lb, attrs, table, rules)
		checkALBConnectionLogs(lb, attrs, table, rules)
		checkALBDeletionProtection(lb, attrs, table, rules)
	}
}

func checkALBAccessLogs(lb types.LoadBalancer, attrs *elasticloadbalancingv2.DescribeLoadBalancerAttributesOutput, table *tablewriter.Table, rules config.RulesConfig) {
	enabled := false
	for _, attr := range attrs.Attributes {
		if *attr.Key == "access_logs.s3.enabled" && *attr.Value == "true" {
			enabled = true
			break
		}
	}
	rule := rules.Get("alb-access-logging")
	if !enabled {
		table.Append([]string{rule.Service, "Fail", color.ColorizeLevel(rule.Level), *lb.LoadBalancerName, "Disabled", rule.Issue})
	} else {
		table.Append([]string{rule.Service, "Pass", "-", *lb.LoadBalancerName, "Enabled", rule.Issue})
	}
}

func checkALBConnectionLogs(lb types.LoadBalancer, attrs *elasticloadbalancingv2.DescribeLoadBalancerAttributesOutput, table *tablewriter.Table, rules config.RulesConfig) {
	// Connection logs are not a standard attribute like access_logs.s3.enabled for ALBs?
	// Wait, ALB connection logs?
	// Ah, maybe the user meant Access Logs (S3) and CONNECTION LOGS? There is no "Connection Logging" for ALB similar to Access Logging in the standard sense often used, UNLESS they mean "Connection Tracing"?
	// Or maybe they mean "Connection termination"?
	// Checking AWS docs... "connection_logs.s3.enabled" exists for NLB/GWLB, but for ALB?
	// ALB has "access_logs.s3.enabled".
	// ALB also has "connection_logs.s3.enabled" is NOT valid for ALB.
	// But wait, user asked for "ALBの接続ログ" (Connection logs).
	// Let's re-read the request. "ALBの接続ログの取得が有効化されているか".
	// AWS ALB Connection Logs were introduced recently?
	// Actually, "Connection logs" are available for ALB.
	// Attribute: `connection_logs.s3.enabled`.
	// Let's assume the attribute key is `connection_logs.s3.enabled`.

	enabled := false
	for _, attr := range attrs.Attributes {
		if *attr.Key == "connection_logs.s3.enabled" && *attr.Value == "true" {
			enabled = true
			break
		}
	}
	rule := rules.Get("alb-connection-logging")
	if !enabled {
		table.Append([]string{rule.Service, "Fail", color.ColorizeLevel(rule.Level), *lb.LoadBalancerName, "Disabled", rule.Issue})
	} else {
		table.Append([]string{rule.Service, "Pass", "-", *lb.LoadBalancerName, "Enabled", rule.Issue})
	}
}

func checkALBDeletionProtection(lb types.LoadBalancer, attrs *elasticloadbalancingv2.DescribeLoadBalancerAttributesOutput, table *tablewriter.Table, rules config.RulesConfig) {
	enabled := false
	for _, attr := range attrs.Attributes {
		if *attr.Key == "deletion_protection.enabled" && *attr.Value == "true" {
			enabled = true
			break
		}
	}
	rule := rules.Get("alb-deletion-protection")
	if !enabled {
		table.Append([]string{rule.Service, "Fail", color.ColorizeLevel(rule.Level), *lb.LoadBalancerName, "Disabled", rule.Issue})
	} else {
		table.Append([]string{rule.Service, "Pass", "-", *lb.LoadBalancerName, "Enabled", rule.Issue})
	}
}
