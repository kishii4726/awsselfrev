/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// ec2Cmd represents the ec2 command
var ec2Cmd = &cobra.Command{
	Use:   "ec2",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			log.Fatalf("unable to load SDK config, %v", err)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Service", "LEVEL", "MESSAGE"})
		client := ec2.NewFromConfig(cfg)

		// EBSの暗号化がデフォルト有効になっているか確認
		resp, err := client.GetEbsEncryptionByDefault(context.TODO(), &ec2.GetEbsEncryptionByDefaultInput{})
		if err != nil {
			log.Fatalf("%v", err)
		}
		if *resp.EbsEncryptionByDefault == false {
			table.Append([]string{"EC2", "Warning", "EBSのデフォルトの暗号化が有効になっていません"})
		}

		resp2, err := client.DescribeVolumes(context.TODO(), &ec2.DescribeVolumesInput{})
		if err != nil {
			log.Fatalf("%v", err)
		}
		for _, v := range resp2.Volumes {
			if *v.Encrypted == false {
				table.Append([]string{"EC2", "Alert", *v.VolumeId + "が暗号化されていません"})
			}
		}

		resp3, err := client.DescribeSnapshots(context.TODO(), &ec2.DescribeSnapshotsInput{
			OwnerIds: []string{"self"},
		})
		if err != nil {
			log.Fatalf("%v", err)
		}
		for _, v := range resp3.Snapshots {
			// snapshotの暗号化確認
			if *v.Encrypted == false {
				table.Append([]string{"EC2", "Alert", *v.SnapshotId + "が暗号化されていません"})
			}
		}
		table.Render()
	},
}

func init() {
	rootCmd.AddCommand(ec2Cmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ec2Cmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// ec2Cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
