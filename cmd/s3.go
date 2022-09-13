package cmd

import (
	services3 "aws-tacit-knowledge/pkg/aws/service/s3"
	"aws-tacit-knowledge/pkg/color"
	"aws-tacit-knowledge/pkg/config"
	"aws-tacit-knowledge/pkg/table"

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

		for _, bucket := range services3.ListBuckets(client) {
			// 暗号化確認
			if services3.IsBucketEncrypted(client, bucket) == false {
				table.Append([]string{"S3", level_alert, bucket + "が暗号化されていません"})
			}
			// パブリックアクセスブロック確認
			if services3.IsBlockPublicAccessEnabled(client, bucket) == false {
				table.Append([]string{"S3", level_warning, bucket + "のパブリックブロックアクセスがすべてオフになっています"})
			}
			// バケット名に`log`が含まれるバケットにライフサイクルルールが設定されているか確認
			if services3.IsLifeCycleRuleConfiguredLogBucket(client, bucket) == false {
				table.Append([]string{"S3", level_warning, bucket + "にライフサイクルルールが設定されていません"})
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
