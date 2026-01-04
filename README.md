# awsselfrev

CLI tool to check AWS resources against personal best practices.

## Install

### Binary
Download the binary from [Releases](https://github.com/kishii4726/awsselfrev/releases).

```bash
# Example for macOS (arm64)
curl -L https://github.com/kishii4726/awsselfrev/releases/latest/download/awsselfrev_Darwin_arm64.tar.gz | tar xz
chmod +x awsselfrev
sudo mv awsselfrev /usr/local/bin/
```

### From Source
```bash
go build -o awsselfrev main.go
```

## Usage

```bash
# Check all services
awsselfrev all

# Check specific service
awsselfrev s3

# Show only failed checks
awsselfrev all --fail-only (or -f)
```

### Example Output
```text
Executing on AWS Account: 123456789012
+---------+--------+---------+-----------------+----------+--------------------------------+
| SERVICE | STATUS |  LEVEL  |    RESOURCE     | SETTING  |             ISSUE              |
+---------+--------+---------+-----------------+----------+--------------------------------+
| S3      | Fail   | Alert   | my-open-bucket  | Disabled | Block public access is all off |
| S3      | Pass   | -       | my-safe-bucket  | Enabled  | Block public access is all off |
| RDS     | Fail   | Warning | my-db-instance  | Disabled | Delete protection is not enabled|
+---------+--------+---------+-----------------+----------+--------------------------------+
```

## Supported Checks

| Service | Level | Check |
| --- | --- | --- |
| **S3** | Alert | Bucket encryption, Block public access |
| | Warning | Lifecycle policy, Object Lock, SSE-KMS encryption, Server access logging, S3 Storage Lens |
| **EC2** | Warning | Default EBS encryption |
| | Alert | EBS Volume encryption, EBS Snapshot encryption |
| **RDS** | Alert | Storage encryption, Public accessibility, Default parameter group |
| | Warning | Delete protection, Log export, Backup enabled, Auto minor version upgrade, Performance Insights, Maintenance window (22:00-05:00 JST), General/Audit/Error/Slow query logs |
| **VPC** | Info | Name tag, Custom flow log format |
| | Warning | DNS hostname, DNS support, VPC flow logs |
| **ELB** | Alert | Target health |
| | Warning | Access logging, Connection logging, Deletion protection |
| **CloudFront** | Warning | Logging enabled |
| **CloudWatch** | Alert | Log retention (Never expire) |
| | Warning | Log group KMS encryption |
| **ECS** | Alert | Sensitive environment variables |
| | Warning | Container Insights, Circuit breaker, CPU Architecture (ARM64), Propagate tags, ECS Exec logging |
| **ECR** | Warning | Tags immutability, Image scanning |
| | Info | Lifecycle policy |
| **Observability**| Warning | Telemetry resource tags |
| **Route53** | Warning | Query logging |
| **WAFV2** | Warning | Logging enabled |

