/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"log"

	"aws-tacit-knowledge/pkg/color"
	"aws-tacit-knowledge/pkg/config"
	"aws-tacit-knowledge/pkg/table"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/spf13/cobra"
)

// ecrCmd represents the ecr command
var ecrCmd = &cobra.Command{
	Use:   "ecr",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		table := table.SetTable()
		client := ecr.NewFromConfig(cfg)
		_, level_warning, _ := color.SetLevelColor()

		resp, err := client.DescribeRepositories(context.TODO(), &ecr.DescribeRepositoriesInput{
			MaxResults: aws.Int32(100),
		})
		if err != nil {
			log.Fatalf("%v", err)
		}
		for _, v := range *&resp.Repositories {
			// tagのimmutabilityの確認
			if *&v.ImageTagMutability == "MUTABLE" {
				table.Append([]string{"ECR", level_warning, "タグの上書きが可能です(MUTABLE)"})
			}
			// imageのscanの確認
			if *&v.ImageScanningConfiguration.ScanOnPush == false {
				table.Append([]string{"ECR", level_warning, "イメージのスキャンが有効になっていません"})
			}

		}
		table.Render()
	},
}

func init() {
	rootCmd.AddCommand(ecrCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ecrCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// ecrCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
