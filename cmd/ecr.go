/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"errors"
	"log"

	"awsselfrev/pkg/color"
	"awsselfrev/pkg/config"
	"awsselfrev/pkg/table"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
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
		// level_info, level_warning, level_alert := color.SetLevelColor()
		level_info, level_warning, _ := color.SetLevelColor()

		resp, err := client.DescribeRepositories(context.TODO(), &ecr.DescribeRepositoriesInput{
			MaxResults: aws.Int32(100),
		})
		if err != nil {
			log.Fatalf("%v", err)
		}
		for _, v := range *&resp.Repositories {
			// tagのimmutabilityの確認
			if *&v.ImageTagMutability == "MUTABLE" {
				table.Append([]string{"ECR", level_warning, *v.RepositoryName + "でタグの上書きが可能です(MUTABLE)"})
			}
			// imageのscanの確認
			if *&v.ImageScanningConfiguration.ScanOnPush == false {
				table.Append([]string{"ECR", level_warning, *v.RepositoryName + "のイメージのスキャンが有効になっていません"})
			}
			//ライフサイクルポリシーの確認
			_, err := client.GetLifecyclePolicy(context.TODO(), &ecr.GetLifecyclePolicyInput{
				RepositoryName: v.RepositoryName,
			})
			if err != nil {
				var re *awshttp.ResponseError
				if errors.As(err, &re) {
					if re.HTTPStatusCode() == 400 {
						table.Append([]string{"ECR", level_info, *v.RepositoryName + "にライフサイクルポリシーが設定されていません"})
					} else {
						log.Fatalf("%v", err)
					}
				}
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
