package service

import (
	"awsselfrev/internal/aws/api"
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
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

func IsDnsHostnamesEnabled(client api.EC2Client, vpcID string) (bool, error) {
	resp, err := client.DescribeVpcAttribute(context.TODO(), &ec2.DescribeVpcAttributeInput{
		VpcId:     &vpcID,
		Attribute: "enableDnsHostnames",
	})
	if err != nil {
		return false, err
	}
	return *resp.EnableDnsHostnames.Value, nil
}

func IsDnsSupportEnabled(client api.EC2Client, vpcID string) (bool, error) {
	resp, err := client.DescribeVpcAttribute(context.TODO(), &ec2.DescribeVpcAttributeInput{
		VpcId:     &vpcID,
		Attribute: "enableDnsSupport",
	})
	if err != nil {
		return false, err
	}
	return *resp.EnableDnsSupport.Value, nil
}

func IsVpcFlowLogsEnabled(client api.EC2Client, vpcID string) (bool, error) {
	resp, err := client.DescribeFlowLogs(context.TODO(), &ec2.DescribeFlowLogsInput{
		Filter: []types.Filter{
			{
				Name:   aws.String("resource-id"),
				Values: []string{vpcID},
			},
		},
	})
	if err != nil {
		return false, err
	}
	return len(resp.FlowLogs) > 0, nil
}

func HasCustomFlowLogFormat(client api.EC2Client, vpcID string) bool {
	resp, err := client.DescribeFlowLogs(context.TODO(), &ec2.DescribeFlowLogsInput{
		Filter: []types.Filter{
			{
				Name:   aws.String("resource-id"),
				Values: []string{vpcID},
			},
		},
	})
	if err != nil {
		return false
	}

	for _, fl := range resp.FlowLogs {
		if fl.LogFormat != nil {
			f := *fl.LogFormat
			if strings.Contains(f, "tcp-flags") &&
				strings.Contains(f, "pkt-srcaddr") &&
				strings.Contains(f, "pkt-dstaddr") &&
				strings.Contains(f, "flow-direction") {
				return true
			}
		}
	}
	return false
}
