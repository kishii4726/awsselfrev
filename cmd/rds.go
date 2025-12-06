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
		rules := config.LoadRules()
		tbl := table.SetTable()
		client := rds.NewFromConfig(cfg)

		checkRDSConfigurations(client, tbl, rules)

		table.Render("RDS", tbl)
	},
}

func checkRDSConfigurations(client *rds.Client, table *tablewriter.Table, rules config.RulesConfig) {
	resp, err := client.DescribeDBClusters(context.TODO(), &rds.DescribeDBClustersInput{})
	if err != nil {
		log.Fatalf("Failed to describe DB clusters: %v", err)
	}

	for _, cluster := range resp.DBClusters {
		checkStorageEncryption(cluster, table, rules)
		checkDeletionProtection(cluster, table, rules)
		checkLogExports(cluster, table, rules)
		checkDBInstances(client, cluster.DBClusterMembers, table, rules)
	}
}

func checkStorageEncryption(cluster types.DBCluster, table *tablewriter.Table, rules config.RulesConfig) {
	if cluster.StorageEncrypted != nil && !*cluster.StorageEncrypted {
		rule := rules.Get("rds-storage-encryption")
		table.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), *cluster.DBClusterIdentifier, rule.Issue})
	}
}

func checkDeletionProtection(cluster types.DBCluster, table *tablewriter.Table, rules config.RulesConfig) {
	if cluster.DeletionProtection != nil && !*cluster.DeletionProtection {
		rule := rules.Get("rds-deletion-protection")
		table.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), *cluster.DBClusterIdentifier, rule.Issue})
	}
}

// TODO:ログ種別ごとに確認できるようにする
func checkLogExports(cluster types.DBCluster, table *tablewriter.Table, rules config.RulesConfig) {
	if len(cluster.EnabledCloudwatchLogsExports) == 0 {
		rule := rules.Get("rds-log-export")
		table.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), *cluster.DBClusterIdentifier, rule.Issue})
	}
}

func checkDBInstances(client *rds.Client, members []types.DBClusterMember, table *tablewriter.Table, rules config.RulesConfig) {
	for _, member := range members {
		resp, err := client.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: member.DBInstanceIdentifier,
		})
		if err != nil {
			log.Fatalf("Failed to describe DB instances: %v", err)
		}

		for _, instance := range resp.DBInstances {
			checkAutoMinorVersionUpgrade(instance, table, rules)
		}
	}
}

func checkAutoMinorVersionUpgrade(instance types.DBInstance, table *tablewriter.Table, rules config.RulesConfig) {
	if instance.AutoMinorVersionUpgrade != nil && *instance.AutoMinorVersionUpgrade {
		rule := rules.Get("rds-auto-minor-version-upgrade")
		table.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), *instance.DBInstanceIdentifier, rule.Issue})
	}
}

func init() {
	rootCmd.AddCommand(rdsCmd)
}
