package cmd

import (
	"context"

	"awsselfrev/internal/aws/api"
	ec2Internal "awsselfrev/internal/aws/service/ec2"
	"awsselfrev/internal/color"
	"awsselfrev/internal/config"
	"awsselfrev/internal/table"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/olekukonko/tablewriter"
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

		checkEC2Configurations(client, tbl, rules)

		table.Render("EC2", tbl)
	},
}

func checkEC2Configurations(client api.EC2Client, tbl *tablewriter.Table, rules config.RulesConfig) {
	// 1. EBS Default Encryption
	ebsEncryptionEnabled, err := ec2Internal.IsEbsDefaultEncryptionEnabled(client)
	if err != nil {
		log.Fatalf("Failed to check EBS default encryption: %v", err)
	}
	ruleEbs := rules.Get("ec2-ebs-default-encryption")
	if !ebsEncryptionEnabled {
		tbl.Append([]string{ruleEbs.Service, "Fail", color.ColorizeLevel(ruleEbs.Level), "-", "Disabled", ruleEbs.Issue})
	} else {
		tbl.Append([]string{ruleEbs.Service, "Pass", "-", "-", "Enabled", ruleEbs.Issue})
	}

	// 2. Volume Encryption
	volumesResp, err := client.DescribeVolumes(context.TODO(), &ec2.DescribeVolumesInput{})
	if err != nil {
		log.Fatalf("Failed to describe volumes: %v", err)
	}
	ruleVol := rules.Get("ec2-volume-encryption")
	if len(volumesResp.Volumes) == 0 {
		tbl.Append([]string{ruleVol.Service, "Pass", "-", "No volumes", "-", ruleVol.Issue})
	} else {
		for _, v := range volumesResp.Volumes {
			if !*v.Encrypted {
				tbl.Append([]string{ruleVol.Service, "Fail", color.ColorizeLevel(ruleVol.Level), *v.VolumeId, "Disabled", ruleVol.Issue})
			} else {
				tbl.Append([]string{ruleVol.Service, "Pass", "-", *v.VolumeId, "Enabled", ruleVol.Issue})
			}
		}
	}

	// 3. Snapshot Encryption
	snapshotsResp, err := client.DescribeSnapshots(context.TODO(), &ec2.DescribeSnapshotsInput{
		OwnerIds: []string{"self"},
	})
	if err != nil {
		log.Fatalf("Failed to describe snapshots: %v", err)
	}
	ruleSnap := rules.Get("ec2-snapshot-encryption")
	if len(snapshotsResp.Snapshots) == 0 {
		tbl.Append([]string{ruleSnap.Service, "Pass", "-", "No snapshots", "-", ruleSnap.Issue})
	} else {
		for _, s := range snapshotsResp.Snapshots {
			if !*s.Encrypted {
				tbl.Append([]string{ruleSnap.Service, "Fail", color.ColorizeLevel(ruleSnap.Level), *s.SnapshotId, "Disabled", ruleSnap.Issue})
			} else {
				tbl.Append([]string{ruleSnap.Service, "Pass", "-", *s.SnapshotId, "Enabled", ruleSnap.Issue})
			}
		}
	}
}

func init() {
	rootCmd.AddCommand(ec2Cmd)
}
