package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// vpcCmd represents the vpc command
var vpcCmd = &cobra.Command{
	Use:   "vpc",
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
		table.SetHeader([]string{"LEVEL", "MESSAGE"})
		fmt.Println("VPC: Check Results")

		client := ec2.NewFromConfig(cfg)
		resp, err := client.DescribeVpcs(context.TODO(), &ec2.DescribeVpcsInput{})
		var data [][]string
		for _, v := range resp.Vpcs {
			// Nameタグの存在確認
			vpc_id := *v.VpcId
			var is_name_tag_exists bool
			for _, v := range *&v.Tags {
				if *v.Key == "Name" {
					is_name_tag_exists = true
				}
			}
			if is_name_tag_exists == false {
				data := append(data, []string{"Notice", vpc_id + "にNameタグが設定されていません"})
				for _, v := range data {
					table.Append(v)
				}
				// fmt.Println("[Notice]:  " + vpc_id + "にNameタグが設定されていません")
			}

			// DNSホスト名の有効確認
			enable_dns_hostnames, err := client.DescribeVpcAttribute(context.TODO(), &ec2.DescribeVpcAttributeInput{
				VpcId:     &vpc_id,
				Attribute: "enableDnsHostnames",
			})
			if err != nil {
				log.Fatalf("%v", err)
			}
			if *enable_dns_hostnames.EnableDnsHostnames.Value == false {
				data := append(data, []string{"Warning", vpc_id + "のDNSホスト名が無効になっています"})
				for _, v := range data {
					table.Append(v)
				}
				// fmt.Println("[Warning]: " + vpc_id + "のDNSホスト名が無効になっています")
			}

			// DNS解決の有効確認
			enable_dns_support, err := client.DescribeVpcAttribute(context.TODO(), &ec2.DescribeVpcAttributeInput{
				VpcId:     &vpc_id,
				Attribute: "enableDnsSupport",
			})
			if err != nil {
				log.Fatalf("%v", err)
			}
			if *enable_dns_support.EnableDnsSupport.Value == false {
				data := append(data, []string{"Warning", vpc_id + "のDNS解決が無効になっています"})
				for _, v := range data {
					table.Append(v)
				}
				// fmt.Println("[Warning]: " + vpc_id + "のDNS解決が無効になっています")
			}
		}
		table.Render()
		// fmt.Println("VPC: Check Completed")
	},
}

func init() {
	rootCmd.AddCommand(vpcCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// vpcCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// vpcCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
