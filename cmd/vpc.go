package cmd

import (
	"context"
	"log"

	"awsselfrev/internal/aws/api"
	ec2Internal "awsselfrev/internal/aws/service/ec2"
	"awsselfrev/internal/color"
	"awsselfrev/internal/config"
	"awsselfrev/internal/table"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var vpcCmd = &cobra.Command{
	Use:   "vpc",
	Short: "Describe and check VPC attributes",
	Long: `The "vpc" command allows you to describe and check various attributes of your VPCs.

This command retrieves information about your VPCs and checks for the presence of the "Name" tag,
as well as the status of DNS hostnames and DNS support. It also checks if VPC Flow Logs are enabled.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		rules := config.LoadRules()
		tbl := table.SetTable()
		client := ec2.NewFromConfig(cfg)
		_, _, _ = color.SetLevelColor()

		checkVPCConfigurations(client, tbl, rules)

		table.Render("VPC", tbl)
	},
}

func checkVPCConfigurations(client api.EC2Client, tbl *tablewriter.Table, rules config.RulesConfig) {
	resp, err := client.DescribeVpcs(context.TODO(), &ec2.DescribeVpcsInput{})
	if err != nil {
		log.Fatalf("Failed to describe VPCs: %v", err)
	}

	if len(resp.Vpcs) == 0 {
		table.AddRow(tbl, []string{"VPC", "-", "-", "No VPCs", "-", "-"})
		return
	}

	for _, vpc := range resp.Vpcs {
		vpcID := *vpc.VpcId
		name := "Missing"
		for _, tag := range vpc.Tags {
			if *tag.Key == "Name" {
				name = *tag.Value
				break
			}
		}

		// 1. Name Tag
		ruleName := rules.Get("vpc-name-tag")
		if name == "Missing" {
			table.AddRow(tbl, []string{ruleName.Service, "Fail", color.ColorizeLevel(ruleName.Level), vpcID, name, ruleName.Issue})
		} else {
			table.AddRow(tbl, []string{ruleName.Service, "Pass", "-", vpcID, name, ruleName.Issue})
		}

		// 2. DNS Hostname
		dnsHostnameEnabled, err := ec2Internal.IsDnsHostnamesEnabled(client, vpcID)
		if err != nil {
			log.Fatalf("Failed to check DNS hostname for VPC %s: %v", vpcID, err)
		}
		ruleDnsH := rules.Get("vpc-dns-hostname")
		if !dnsHostnameEnabled {
			table.AddRow(tbl, []string{ruleDnsH.Service, "Fail", color.ColorizeLevel(ruleDnsH.Level), vpcID, "Disabled", ruleDnsH.Issue})
		} else {
			table.AddRow(tbl, []string{ruleDnsH.Service, "Pass", "-", vpcID, "Enabled", ruleDnsH.Issue})
		}

		// 3. DNS Support
		dnsSupportEnabled, err := ec2Internal.IsDnsSupportEnabled(client, vpcID)
		if err != nil {
			log.Fatalf("Failed to check DNS support for VPC %s: %v", vpcID, err)
		}
		ruleDnsS := rules.Get("vpc-dns-support")
		if !dnsSupportEnabled {
			table.AddRow(tbl, []string{ruleDnsS.Service, "Fail", color.ColorizeLevel(ruleDnsS.Level), vpcID, "Disabled", ruleDnsS.Issue})
		} else {
			table.AddRow(tbl, []string{ruleDnsS.Service, "Pass", "-", vpcID, "Enabled", ruleDnsS.Issue})
		}

		// 4. Flow Logs
		flowLogsEnabled, err := ec2Internal.IsVpcFlowLogsEnabled(client, vpcID)
		if err != nil {
			log.Fatalf("Failed to check Flow Logs for VPC %s: %v", vpcID, err)
		}
		ruleFlow := rules.Get("vpc-flow-logs")
		if !flowLogsEnabled {
			table.AddRow(tbl, []string{ruleFlow.Service, "Fail", color.ColorizeLevel(ruleFlow.Level), vpcID, "Disabled", ruleFlow.Issue})
		} else {
			// Flow logs enabled, check custom format
			ruleFormat := rules.Get("vpc-flow-logs-custom-format")
			if !ec2Internal.HasCustomFlowLogFormat(client, vpcID) { // Using new internal function
				table.AddRow(tbl, []string{ruleFormat.Service, "Fail", color.ColorizeLevel(ruleFormat.Level), vpcID, "Invalid", ruleFormat.Issue})
			} else {
				table.AddRow(tbl, []string{ruleFormat.Service, "Pass", "-", vpcID, "Valid", ruleFormat.Issue})
			}
			// Also report flow logs enabled as Pass
			table.AddRow(tbl, []string{ruleFlow.Service, "Pass", "-", vpcID, "Enabled", ruleFlow.Issue})
		}
	}
}

func init() {
	rootCmd.AddCommand(vpcCmd)
}
