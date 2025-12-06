package service

import (
	"awsselfrev/internal/aws/api"
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func IsEbsDefaultEncryptionEnabled(client api.EC2Client) (bool, error) {
	resp, err := client.GetEbsEncryptionByDefault(context.TODO(), &ec2.GetEbsEncryptionByDefaultInput{})
	if err != nil {
		return false, err
	}
	return *resp.EbsEncryptionByDefault, nil
}

func IsVolumeEncrypted(client api.EC2Client) ([]string, error) {
	var unencryptedVolumes []string

	resp, err := client.DescribeVolumes(context.TODO(), &ec2.DescribeVolumesInput{})
	if err != nil {
		return nil, err
	}

	for _, v := range resp.Volumes {
		if !*v.Encrypted {
			unencryptedVolumes = append(unencryptedVolumes, *v.VolumeId)
		}
	}
	return unencryptedVolumes, nil
}

func IsSnapshotEncrypted(client api.EC2Client) ([]string, error) {
	var snapshotIDs []string

	resp, err := client.DescribeSnapshots(context.TODO(), &ec2.DescribeSnapshotsInput{
		OwnerIds: []string{"self"},
	})
	if err != nil {
		return nil, err
	}

	for _, snapshot := range resp.Snapshots {
		snapshotIDs = append(snapshotIDs, *snapshot.SnapshotId)
	}
	return snapshotIDs, nil
}

func HandleServiceError(err error) bool {
	if err != nil {
		log.Println("Service error:", err)
		return false
	}
	return true
}
