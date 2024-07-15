package cmd

import (
	"context"
	"log"

	"awsselfrev/internal/config"
	"awsselfrev/internal/table"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var rdsCmd = &cobra.Command{
	Use:   "rds",
	Short: "Checks RDS configurations for best practices",
	Long: `This command checks various RDS configurations and best practices such as:
- Storage encryption
- Deletion protection
- Log exports`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		tbl := table.SetTable()
		client := rds.NewFromConfig(cfg)

		checkRDSConfigurations(client, tbl)

		table.Render("RDS", tbl)
	},
}

func checkRDSConfigurations(client *rds.Client, table *tablewriter.Table) {
	resp, err := client.DescribeDBClusters(context.TODO(), &rds.DescribeDBClustersInput{})
	if err != nil {
		log.Fatalf("Failed to describe DB clusters: %v", err)
	}

	for _, cluster := range resp.DBClusters {
		checkStorageEncryption(cluster, table)
		checkDeletionProtection(cluster, table)
		checkLogExports(cluster, table)
		checkDBInstances(client, cluster.DBClusterMembers, table)
	}
}

func checkStorageEncryption(cluster types.DBCluster, table *tablewriter.Table) {
	if cluster.StorageEncrypted != nil && !*cluster.StorageEncrypted {
		table.Append([]string{"RDS", "Alert", *cluster.DBClusterIdentifier + "のStorageが暗号化されていません"})
	}
}

func checkDeletionProtection(cluster types.DBCluster, table *tablewriter.Table) {
	if cluster.DeletionProtection != nil && !*cluster.DeletionProtection {
		table.Append([]string{"RDS", "Warning", *cluster.DBClusterIdentifier + "の削除保護が有効化されていません"})
	}
}

func checkLogExports(cluster types.DBCluster, table *tablewriter.Table) {
	if len(cluster.EnabledCloudwatchLogsExports) == 0 {
		table.Append([]string{"RDS", "Warning", *cluster.DBClusterIdentifier + "でログ出力が設定されていません"})
	}
}

func checkDBInstances(client *rds.Client, members []types.DBClusterMember, table *tablewriter.Table) {
	for _, member := range members {
		resp, err := client.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: member.DBInstanceIdentifier,
		})
		if err != nil {
			log.Fatalf("Failed to describe DB instances: %v", err)
		}

		for _, instance := range resp.DBInstances {
			checkAutoMinorVersionUpgrade(instance, table)
		}
	}
}

func checkAutoMinorVersionUpgrade(instance types.DBInstance, table *tablewriter.Table) {
	if instance.AutoMinorVersionUpgrade != nil && *instance.AutoMinorVersionUpgrade {
		table.Append([]string{"RDS", "Warning", *instance.DBInstanceIdentifier + "のマイナーバージョン自動アップグレードが有効化されています"})
	}
}

func init() {
	rootCmd.AddCommand(rdsCmd)
}
