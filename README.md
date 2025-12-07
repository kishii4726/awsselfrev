# awsselfrev

## Install
```
$ curl -OL <Release assets url>

$ tar -zxvf <Download file name>

$ sudo mv awsselfrev /usr/local/bin
```

## Usage
```
$ awsselfrev ecr
+---------+---------+----------+-------------------------------+
| SERVICE |  LEVEL  | RESOURCE |             ISSUE             |
+---------+---------+----------+-------------------------------+
| ECR     | Warning | repo1    | Image scanning is not enabled |
+---------+---------+----------+-------------------------------+
| ECR     | INFO    | repo2    | Lifecycle policy is not set   |
+---------+---------+----------+-------------------------------+

$ awsselfrev s3
+---------+---------+----------------------------------------------------------------+--------------------------------+
| SERVICE |  LEVEL  |                            RESOURCE                            |             ISSUE              |
+---------+---------+----------------------------------------------------------------+--------------------------------+
| S3      | Alert   | bucket1                                                        | Block public access is all off |
+---------+---------+----------------------------------------------------------------+--------------------------------+
| S3      | Warning | bucket2                                                        | Lifecycle policy is not set    |
+---------+---------+----------------------------------------------------------------+--------------------------------+
```

## Checks
| Service | Level | Check |
| --- | --- | --- |
| S3 | Alert | Bucket encryption is not set |
| S3 | Alert | Block public access is all off |
| S3 | Warning | Lifecycle policy is not set |
| S3 | Warning | Object Lock is not enabled |
| S3 | Warning | SSE-KMS encryption is not set |
| EC2 | Warning | Default encryption for EBS is not set |
| EC2 | Alert | EBS encryption is not set |
| EC2 | Alert | EBS encryption is not set (Snapshot) |
| RDS | Alert | Storage encryption is not set |
| RDS | Warning | Delete protection is not enabled |
| RDS | Warning | Log export is not set |
| RDS | Warning | Auto minor version upgrade is enabled |
| VPC | Info | Name tag is not set |
| VPC | Warning | DNS hostname is not enabled |
| VPC | Warning | DNS support is not enabled |
| VPC | Warning | VPC flow logs is not enabled |
| VPC | Info | Custom flow log format is not set or missing required fields |
| CloudWatchLogs | Alert | Retention is set to never expire |
| ECR | Warning | Tags can be overwritten |
| ECR | Warning | Image scanning is not enabled |
| ECR | Info | Lifecycle policy is not set |
