package cmd

import (
	"awsselfrev/internal/color"
	"awsselfrev/internal/config"
	"awsselfrev/internal/table"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/spf13/cobra"
)

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Execute all commands and combine output",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		rules := config.LoadRules()
		tbl := table.SetTable()
		_, _, _ = color.SetLevelColor()

		// Initialize Clients
		albClient := elasticloadbalancingv2.NewFromConfig(cfg)
		cfClient := cloudfront.NewFromConfig(cfg)
		cwLogsClient := cloudwatchlogs.NewFromConfig(cfg)
		ec2Client := ec2.NewFromConfig(cfg)
		ecrClient := ecr.NewFromConfig(cfg)
		ecsClient := ecs.NewFromConfig(cfg)
		obsClient := observabilityadmin.NewFromConfig(cfg)
		rdsClient := rds.NewFromConfig(cfg)
		route53Client := route53.NewFromConfig(cfg)
		s3Client := s3.NewFromConfig(cfg)

		// Run Checks
		checkALBConfigurations(albClient, tbl, rules)
		checkCloudFrontConfigurations(cfClient, tbl, rules)
		checkCloudWatchLogsConfigurations(cwLogsClient, tbl, rules)
		checkEC2Configurations(ec2Client, tbl, rules)
		checkECRConfigurations(ecrClient, tbl, rules)
		checkECSConfigurations(ecsClient, tbl, rules)
		checkObservabilityConfigurations(obsClient, tbl, rules)
		checkRDSConfigurations(rdsClient, tbl, rules)
		checkRoute53Configurations(route53Client, tbl, rules)
		checkS3Configurations(s3Client, tbl, rules)
		checkVPCConfigurations(ec2Client, tbl, rules)

		table.Render("All Services", tbl)
	},
}

func init() {
	rootCmd.AddCommand(allCmd)
}
