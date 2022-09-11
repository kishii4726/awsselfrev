package service

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func ConfirmEbsEncryptionByDefault(client *ec2.Client) bool {
	resp, err := client.GetEbsEncryptionByDefault(context.TODO(), &ec2.GetEbsEncryptionByDefaultInput{})
	if err != nil {
		log.Fatalf("%v", err)
	}
	return *resp.EbsEncryptionByDefault
}

func ConfirmVolumeEncryption(client *ec2.Client) []string {
	var l []string

	resp, err := client.DescribeVolumes(context.TODO(), &ec2.DescribeVolumesInput{})
	if err != nil {
		log.Fatalf("%v", err)
	}
	for _, v := range resp.Volumes {
		if *v.Encrypted == false {
			l = append(l, *v.VolumeId)
		}
	}
	return l
}

func ConfirmSnapshotEncryption(client *ec2.Client) []string {
	var l []string

	resp, err := client.DescribeSnapshots(context.TODO(), &ec2.DescribeSnapshotsInput{
		OwnerIds: []string{"self"},
	})
	if err != nil {
		for _, v := range resp.Snapshots {
			if *v.Encrypted == false {
				l = append(l, *v.SnapshotId)
			}
		}
	}
	return l
}
