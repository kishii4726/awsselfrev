package cmd

import (
	"context"
	"log"

	"awsselfrev/internal/color"
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

		_, levelWarning, levelAlert := color.SetLevelColor()

		checkRDSConfigurations(client, tbl, levelWarning, levelAlert)

		table.Render("RDS", tbl)
	},
}

func checkRDSConfigurations(client *rds.Client, table *tablewriter.Table, levelWarning string, levelAlert string) {
	resp, err := client.DescribeDBClusters(context.TODO(), &rds.DescribeDBClustersInput{})
	if err != nil {
		log.Fatalf("Failed to describe DB clusters: %v", err)
	}

	for _, cluster := range resp.DBClusters {
		checkStorageEncryption(cluster, table, levelAlert)
		checkDeletionProtection(cluster, table, levelWarning)
		checkLogExports(cluster, table, levelWarning)
		checkDBInstances(client, cluster.DBClusterMembers, table, levelWarning)
	}
}

func checkStorageEncryption(cluster types.DBCluster, table *tablewriter.Table, level string) {
	if cluster.StorageEncrypted != nil && !*cluster.StorageEncrypted {
		table.Append([]string{"RDS", level, "Storage encryption is not set", *cluster.DBClusterIdentifier})
	}
}

func checkDeletionProtection(cluster types.DBCluster, table *tablewriter.Table, level string) {
	if cluster.DeletionProtection != nil && !*cluster.DeletionProtection {
		table.Append([]string{"RDS", level, "Delete protection is not enabled", *cluster.DBClusterIdentifier})
	}
}

// TODO:ログ種別ごとに確認できるようにする
func checkLogExports(cluster types.DBCluster, table *tablewriter.Table, level string) {
	if len(cluster.EnabledCloudwatchLogsExports) == 0 {
		table.Append([]string{"RDS", level, "Log export is not set", *cluster.DBClusterIdentifier})
	}
}

func checkDBInstances(client *rds.Client, members []types.DBClusterMember, table *tablewriter.Table, level string) {
	for _, member := range members {
		resp, err := client.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: member.DBInstanceIdentifier,
		})
		if err != nil {
			log.Fatalf("Failed to describe DB instances: %v", err)
		}

		for _, instance := range resp.DBInstances {
			checkAutoMinorVersionUpgrade(instance, table, level)
		}
	}
}

func checkAutoMinorVersionUpgrade(instance types.DBInstance, table *tablewriter.Table, level string) {
	if instance.AutoMinorVersionUpgrade != nil && *instance.AutoMinorVersionUpgrade {
		table.Append([]string{"RDS", level, "Auto minor version upgrade is enabled", *instance.DBInstanceIdentifier})
	}
}

func init() {
	rootCmd.AddCommand(rdsCmd)
}
