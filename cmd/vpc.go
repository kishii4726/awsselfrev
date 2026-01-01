package cmd

import (
	"context"
	"log"
	"strings"

	"awsselfrev/internal/aws/api"
	"awsselfrev/internal/color"
	"awsselfrev/internal/config"
	"awsselfrev/internal/table"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
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

		resp, err := client.DescribeVpcs(context.TODO(), &ec2.DescribeVpcsInput{})
		if err != nil {
			log.Fatalf("Failed to describe VPCs: %v", err)
		}

		var data [][]string
		for _, vpc := range resp.Vpcs {
			vpcID := *vpc.VpcId

			ruleName := rules.Get("vpc-name-tag")
			if !hasNameTag(vpc.Tags) {
				message := []string{ruleName.Service, "NG", color.ColorizeLevel(ruleName.Level), vpcID, ruleName.Issue}
				data = append(data, message)
			} else {
				message := []string{ruleName.Service, "Pass", "-", vpcID, ruleName.Issue}
				data = append(data, message)
			}

			ruleDnsHost := rules.Get("vpc-dns-hostname")
			if !isAttributeEnabled(client, vpcID, types.VpcAttributeNameEnableDnsHostnames) {
				message := []string{ruleDnsHost.Service, "NG", color.ColorizeLevel(ruleDnsHost.Level), vpcID, ruleDnsHost.Issue}
				data = append(data, message)
			} else {
				message := []string{ruleDnsHost.Service, "Pass", "-", vpcID, ruleDnsHost.Issue}
				data = append(data, message)
			}

			ruleDnsSup := rules.Get("vpc-dns-support")
			if !isAttributeEnabled(client, vpcID, types.VpcAttributeNameEnableDnsSupport) {
				message := []string{ruleDnsSup.Service, "NG", color.ColorizeLevel(ruleDnsSup.Level), vpcID, ruleDnsSup.Issue}
				data = append(data, message)
			} else {
				message := []string{ruleDnsSup.Service, "Pass", "-", vpcID, ruleDnsSup.Issue}
				data = append(data, message)
			}

			ruleFlow := rules.Get("vpc-flow-logs")
			if !isFlowLogsEnabled(client, vpcID) {
				message := []string{ruleFlow.Service, "NG", color.ColorizeLevel(ruleFlow.Level), vpcID, ruleFlow.Issue}
				data = append(data, message)
			} else {
				// Flow logs enabled, check custom format
				ruleFormat := rules.Get("vpc-flow-logs-custom-format")
				if !hasCustomFlowLogFormat(client, vpcID) {
					message := []string{ruleFormat.Service, "NG", color.ColorizeLevel(ruleFormat.Level), vpcID, ruleFormat.Issue}
					data = append(data, message)
				} else {
					message := []string{ruleFormat.Service, "Pass", "-", vpcID, ruleFormat.Issue}
					data = append(data, message)
				}
				// Also report flow logs enabled as Pass
				message := []string{ruleFlow.Service, "Pass", "-", vpcID, ruleFlow.Issue}
				data = append(data, message)
			}
		}

		for _, item := range data {
			tbl.Append(item)
		}

		table.Render("VPC", tbl)
	},
}

func hasNameTag(tags []types.Tag) bool {
	for _, tag := range tags {
		if *tag.Key == "Name" {
			return true
		}
	}
	return false
}

func isAttributeEnabled(client api.EC2Client, vpcID string, attribute types.VpcAttributeName) bool {
	resp, err := client.DescribeVpcAttribute(context.TODO(), &ec2.DescribeVpcAttributeInput{
		VpcId:     &vpcID,
		Attribute: attribute,
	})
	if err != nil {
		log.Fatalf("Failed to describe VPC attribute %s for VPC %s: %v", attribute, vpcID, err)
	}
	switch attribute {
	case types.VpcAttributeNameEnableDnsHostnames:
		return *resp.EnableDnsHostnames.Value
	case types.VpcAttributeNameEnableDnsSupport:
		return *resp.EnableDnsSupport.Value
	}
	return false
}

func isFlowLogsEnabled(client api.EC2Client, vpcID string) bool {
	describeFlowLogsInput := &ec2.DescribeFlowLogsInput{
		Filter: []types.Filter{
			{
				Name:   aws.String("resource-id"),
				Values: []string{vpcID},
			},
		},
	}

	resp, err := client.DescribeFlowLogs(context.TODO(), describeFlowLogsInput)
	if err != nil {
		log.Fatalf("Failed to describe flow logs for VPC %s: %v", vpcID, err)
	}

	return len(resp.FlowLogs) > 0
}

func hasCustomFlowLogFormat(client api.EC2Client, vpcID string) bool {
	describeFlowLogsInput := &ec2.DescribeFlowLogsInput{
		Filter: []types.Filter{
			{
				Name:   aws.String("resource-id"),
				Values: []string{vpcID},
			},
		},
	}

	resp, err := client.DescribeFlowLogs(context.TODO(), describeFlowLogsInput)
	if err != nil {
		log.Fatalf("Failed to describe flow logs for VPC %s: %v", vpcID, err)
	}

	for _, fl := range resp.FlowLogs {
		if fl.LogFormat != nil {
			fmt := *fl.LogFormat
			if strings.Contains(fmt, "tcp-flags") &&
				strings.Contains(fmt, "pkt-srcaddr") &&
				strings.Contains(fmt, "pkt-dstaddr") &&
				strings.Contains(fmt, "flow-direction") {
				return true
			}
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(vpcCmd)
}
