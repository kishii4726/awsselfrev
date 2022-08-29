package cmd

import (
	"context"
	"errors"
	"log"
	"strings"

	"aws-tacit-knowledge/pkg/color"
	"aws-tacit-knowledge/pkg/config"
	"aws-tacit-knowledge/pkg/table"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
)

// s3Cmd represents the s3 command
var s3Cmd = &cobra.Command{
	Use:   "s3",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		table := table.SetTable()
		client := s3.NewFromConfig(cfg)
		// level_info, level_warning, level_alert := color.SetLevelColor()
		_, level_warning, level_alert := color.SetLevelColor()

		// s3バケット一覧取得
		var s3_buckets []string
		resp, err := client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
		if err != nil {
			log.Fatalf("%v", err)
		}
		for _, v := range resp.Buckets {
			s3_buckets = append(s3_buckets, *v.Name)
		}

		for _, v := range s3_buckets {
			// 暗号化確認
			_, err := client.GetBucketEncryption(context.TODO(), &s3.GetBucketEncryptionInput{
				Bucket: aws.String(v),
			})
			if err != nil {
				var re *awshttp.ResponseError
				if errors.As(err, &re) {
					if re.HTTPStatusCode() == 404 {
						table.Append([]string{"S3", level_alert, v + "が暗号化されていません"})
					} else if re.HTTPStatusCode() == 301 {
					} else {
						log.Fatalf("%v", err)
					}
				}
			}

			// パブリックアクセスブロック確認
			resp, err := client.GetPublicAccessBlock(context.TODO(), &s3.GetPublicAccessBlockInput{
				Bucket: aws.String(v),
			})
			_ = resp
			if err != nil {
				var re *awshttp.ResponseError
				if errors.As(err, &re) {
					if re.HTTPStatusCode() == 404 {
						table.Append([]string{"S3", level_warning, v + "のパブリックアクセスブロックがすべてオフになっています"})
					} else if re.HTTPStatusCode() == 301 {
					} else {
						log.Fatalf("%v", err)
					}
				}
			}

			// バケット名に`log`が含まれるバケットにライフサイクルルールが設定されているか確認
			if strings.Contains(v, "log") {
				_, err := client.GetBucketLifecycleConfiguration(context.TODO(), &s3.GetBucketLifecycleConfigurationInput{
					Bucket: aws.String(v),
				})
				if err != nil {
					var re *awshttp.ResponseError
					if errors.As(err, &re) {
						if re.HTTPStatusCode() == 404 {
							table.Append([]string{"S3", level_warning, v + "にライフサイクルルールが設定されていません"})
						} else if re.HTTPStatusCode() == 301 {
						} else {
							log.Fatalf("%v", err)
						}
					}
				}
			}
		}

		table.Render()
	},
}

func init() {
	rootCmd.AddCommand(s3Cmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// s3Cmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// s3Cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
