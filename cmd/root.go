package cmd

import (
	"awsselfrev/internal/config"
	"awsselfrev/internal/table"
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/spf13/cobra"
)

var version string

var rootCmd = &cobra.Command{
	Use:   "awsselfrev",
	Short: "Personal AWS best practice checker",
	Long: `awsselfrev is a CLI tool that checks your AWS resource configurations against
personal best practices. It evaluates settings for various services—including
S3, RDS, EC2, and VPC—and provides a consolidated report on their current status.`,
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		client := sts.NewFromConfig(cfg)
		identity, err := client.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
		if err != nil {
			fmt.Printf("Failed to get AWS identity: %v\n", err)
			return
		}
		fmt.Printf("Executing on AWS Account: %s\n", *identity.Account)

		failOnly, _ := cmd.Flags().GetBool("fail-only")
		table.FailOnly = failOnly
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("fail-only", "f", false, "Show only failed checks")
}
