package cmd

import (
	"context"
	"log"

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

			if !hasNameTag(vpc.Tags) {
				rule := rules.Rules["vpc-name-tag"]
				message := []string{rule.Service, color.ColorizeLevel(rule.Level), vpcID, rule.Issue}
				data = append(data, message)
			}

			if !isAttributeEnabled(client, vpcID, types.VpcAttributeNameEnableDnsHostnames) {
				rule := rules.Rules["vpc-dns-hostname"]
				message := []string{rule.Service, color.ColorizeLevel(rule.Level), vpcID, rule.Issue}
				data = append(data, message)
			}

			if !isAttributeEnabled(client, vpcID, types.VpcAttributeNameEnableDnsSupport) {
				rule := rules.Rules["vpc-dns-support"]
				message := []string{rule.Service, color.ColorizeLevel(rule.Level), vpcID, rule.Issue}
				data = append(data, message)
			}

			if !isFlowLogsEnabled(client, vpcID) {
				rule := rules.Rules["vpc-flow-logs"]
				message := []string{rule.Service, color.ColorizeLevel(rule.Level), vpcID, rule.Issue}
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

func isAttributeEnabled(client *ec2.Client, vpcID string, attribute types.VpcAttributeName) bool {
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

func isFlowLogsEnabled(client *ec2.Client, vpcID string) bool {
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

func init() {
	rootCmd.AddCommand(vpcCmd)
}
