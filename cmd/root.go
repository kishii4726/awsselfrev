package cmd

import (
	"awsselfrev/internal/config"
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/spf13/cobra"
)

var version string

var rootCmd = &cobra.Command{
	Use:   "awsselfrev",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
