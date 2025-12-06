package cmd

import (
	ec2Internal "awsselfrev/internal/aws/service/ec2"
	"awsselfrev/internal/color"
	"awsselfrev/internal/config"
	"awsselfrev/internal/table"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
)

var ec2Cmd = &cobra.Command{
	Use:   "ec2",
	Short: "Check EC2 resources for best practices and configurations",
	Long: `This command checks various EC2 configurations and best practices such as:
- EBS default encryption
- Volume encryption
- Snapshot encryption`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		rules := config.LoadRules()
		tbl := table.SetTable()
		client := ec2.NewFromConfig(cfg)
		_, _, _ = color.SetLevelColor()

		ebsEncryptionEnabled, err := ec2Internal.IsEbsDefaultEncryptionEnabled(client)
		if err != nil {
			log.Fatalf("Failed to check EBS default encryption: %v", err)
		}
		if !ebsEncryptionEnabled {
			rule := rules.Rules["ec2-ebs-default-encryption"]
			tbl.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), "-", rule.Issue})
		}

		unencryptedVolumes, err := ec2Internal.IsVolumeEncrypted(client)
		if err != nil {
			log.Fatalf("Failed to check volume encryption: %v", err)
		}
		for _, v := range unencryptedVolumes {
			rule := rules.Rules["ec2-volume-encryption"]
			tbl.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), v, rule.Issue})
		}

		encryptedSnapshots, err := ec2Internal.IsSnapshotEncrypted(client)
		if err != nil {
			log.Fatalf("Failed to check snapshot encryption: %v", err)
		}
		for _, v := range encryptedSnapshots {
			rule := rules.Rules["ec2-snapshot-encryption"]
			tbl.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), v, rule.Issue})
		}

		table.Render("EC2", tbl)
	},
}

func init() {
	rootCmd.AddCommand(ec2Cmd)
}
