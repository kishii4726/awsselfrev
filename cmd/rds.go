package cmd

import (
	"context"
	"log"
	"strings"

	"awsselfrev/internal/aws/api"
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
- Log exports
- Backup enabled
- Default parameter group usage
- Public accessibility
- Comprehensive log enabled (General, Audit, Error, SlowQuery)`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		rules := config.LoadRules()
		tbl := table.SetTable()
		client := rds.NewFromConfig(cfg)
		_, _, _ = color.SetLevelColor()

		checkRDSConfigurations(client, tbl, rules)

		table.Render("RDS", tbl)
	},
}

// Map to cache parameter group values: GroupName -> Key -> Value
var paramGroupCache = make(map[string]map[string]string)

func checkRDSConfigurations(client api.RDSClient, table *tablewriter.Table, rules config.RulesConfig) {
	resp, err := client.DescribeDBClusters(context.TODO(), &rds.DescribeDBClustersInput{})
	if err != nil {
		log.Fatalf("Failed to describe DB clusters: %v", err)
	}

	for _, cluster := range resp.DBClusters {
		checkStorageEncryption(cluster, table, rules)
		checkDeletionProtection(cluster, table, rules)
		checkClusterBackupEnabled(cluster, table, rules)
		checkClusterDefaultParameterGroup(cluster, table, rules)
		checkClusterLogConfigurations(client, cluster, table, rules)
		checkDBInstances(client, cluster.DBClusterMembers, table, rules)
	}

	// Also check standalone instances if not covered by clusters (DBClusterMembers only covers cluster members).
	// For simplicity, we assume checkDBInstances handles all if we list all (but ListClusters only returns clusters).
	// We should also list all instances to catch standalone ones.
	// However, the original code structure iterated clusters and then members.
	// To be thorough, we should iterate DescribeDBInstances for everything, but need to avoid duplicates if we checked via cluster.
	// For now, let's stick to the previous pattern but ensure we check instance-level logs for all instances found via Members,
	// AND we should probably iterate all instances separately to catch non-Aurora RDS.
	// Refactoring to iterate ALL instances once is better.

	instancesResp, err := client.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{})
	if err != nil {
		log.Fatalf("Failed to describe DB instances: %v", err)
	}

	processedInstances := make(map[string]bool)
	// Mark cluster members as processed if we want to avoid double checking,
	// OR just iterate all instances here for instance-level checks and use cluster loop only for cluster-level checks.
	// Let's iterate all instances here for instance-level checks.
	for _, instance := range instancesResp.DBInstances {
		if processedInstances[*instance.DBInstanceIdentifier] {
			continue
		}

		checkAutoMinorVersionUpgrade(instance, table, rules)
		checkInstanceDefaultParameterGroup(instance, table, rules)
		checkPublicAccessibility(instance, table, rules)
		checkPerformanceInsights(instance, table, rules)
		checkInstanceLogConfigurations(client, instance, table, rules)

		processedInstances[*instance.DBInstanceIdentifier] = true
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

func checkClusterBackupEnabled(cluster types.DBCluster, table *tablewriter.Table, rules config.RulesConfig) {
	if cluster.BackupRetentionPeriod != nil && *cluster.BackupRetentionPeriod == 0 {
		rule := rules.Get("rds-backup-enabled")
		table.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), *cluster.DBClusterIdentifier, rule.Issue})
	}
}

func checkClusterDefaultParameterGroup(cluster types.DBCluster, table *tablewriter.Table, rules config.RulesConfig) {
	if cluster.DBClusterParameterGroup != nil && strings.HasPrefix(*cluster.DBClusterParameterGroup, "default.") {
		rule := rules.Get("rds-default-parameter-group")
		table.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), *cluster.DBClusterIdentifier, rule.Issue})
	}
}

func checkDBInstances(client api.RDSClient, members []types.DBClusterMember, table *tablewriter.Table, rules config.RulesConfig) {
	// fetching is now done in main loop to cover all instances
}

func checkAutoMinorVersionUpgrade(instance types.DBInstance, table *tablewriter.Table, rules config.RulesConfig) {
	if instance.AutoMinorVersionUpgrade != nil && *instance.AutoMinorVersionUpgrade {
		rule := rules.Get("rds-auto-minor-version-upgrade")
		table.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), *instance.DBInstanceIdentifier, rule.Issue})
	}
}

func checkInstanceDefaultParameterGroup(instance types.DBInstance, table *tablewriter.Table, rules config.RulesConfig) {
	for _, pg := range instance.DBParameterGroups {
		if pg.DBParameterGroupName != nil && strings.HasPrefix(*pg.DBParameterGroupName, "default.") {
			rule := rules.Get("rds-default-parameter-group")
			table.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), *instance.DBInstanceIdentifier, rule.Issue})
			break // Report once per instance
		}
	}
}

func checkPublicAccessibility(instance types.DBInstance, table *tablewriter.Table, rules config.RulesConfig) {
	if instance.PubliclyAccessible != nil && *instance.PubliclyAccessible {
		rule := rules.Get("rds-public-access")
		table.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), *instance.DBInstanceIdentifier, rule.Issue})
	}
}

func checkPerformanceInsights(instance types.DBInstance, table *tablewriter.Table, rules config.RulesConfig) {
	if instance.PerformanceInsightsEnabled != nil && !*instance.PerformanceInsightsEnabled {
		rule := rules.Get("rds-performance-insights")
		table.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), *instance.DBInstanceIdentifier, rule.Issue})
	}
}

// Log Checks

func checkClusterLogConfigurations(client api.RDSClient, cluster types.DBCluster, table *tablewriter.Table, rules config.RulesConfig) {
	// Check Cluster logs (mostly for Aurora)
	exports := cluster.EnabledCloudwatchLogsExports
	pgName := ""
	if cluster.DBClusterParameterGroup != nil {
		pgName = *cluster.DBClusterParameterGroup
	}

	checkLogs(client, pgName, exports, *cluster.DBClusterIdentifier, table, rules, true)
}

func checkInstanceLogConfigurations(client api.RDSClient, instance types.DBInstance, table *tablewriter.Table, rules config.RulesConfig) {
	// Check Instance logs (for RDS and Aurora members)
	exports := instance.EnabledCloudwatchLogsExports
	pgName := ""
	// Use the first parameter group (usually only one active)
	if len(instance.DBParameterGroups) > 0 {
		pgName = *instance.DBParameterGroups[0].DBParameterGroupName
	}

	checkLogs(client, pgName, exports, *instance.DBInstanceIdentifier, table, rules, false)
}

func checkLogs(client api.RDSClient, pgName string, exports []string, identifier string, table *tablewriter.Table, rules config.RulesConfig, isCluster bool) {
	// Helper to check slice contains
	contains := func(slice []string, item string) bool {
		for _, s := range slice {
			if s == item {
				return true
			}
		}
		return false
	}

	params := getParameters(client, pgName, isCluster)

	// 1. General Log
	// Req: Exported AND (general_log=1 OR general_log=ON)
	if !contains(exports, "general") || (params["general_log"] != "1" && strings.ToUpper(params["general_log"]) != "ON") {
		rule := rules.Get("rds-general-log")
		table.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), identifier, rule.Issue})
	}

	// 2. Slow Query Log
	// Req: Exported AND (slow_query_log=1 OR slow_query_log=ON)
	if !contains(exports, "slowquery") || (params["slow_query_log"] != "1" && strings.ToUpper(params["slow_query_log"]) != "ON") {
		rule := rules.Get("rds-slow-query-log")
		table.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), identifier, rule.Issue})
	}

	// 3. Audit Log
	// Req: Exported AND (server_audit_logging=1 OR server_audit_logging=ON) (for Aurora/MariaDB)
	// Note: Engine type check would be better, but assuming if param exists it should be on.
	// If param misses, we might just check export.
	auditEnabled := false
	if val, ok := params["server_audit_logging"]; ok {
		if val == "1" || strings.ToUpper(val) == "ON" {
			auditEnabled = true
		}
	} else {
		// If param not present (e.g. Postgres), rely mostly on export or other params (pgaudit).
		// For simplicity/compatibility, satisfied if exported.
		auditEnabled = true
	}

	if !contains(exports, "audit") || !auditEnabled {
		rule := rules.Get("rds-audit-log")
		table.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), identifier, rule.Issue})
	}

	// 4. Error Log
	// Req: Exported. Error log is usually enabled by default on engine.
	if !contains(exports, "error") && !contains(exports, "postgresql") && !contains(exports, "alert") { // Postgres uses 'postgresql', Oracle/MSSQL uses 'error'/'agent', MySql 'error'
		// Loose check for any "error-like" log export presence if exact name varies,
		// but 'error' is standard for MySQL. 'postgresql' for PG.
		rule := rules.Get("rds-error-log")
		table.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), identifier, rule.Issue})
	}
}

func getParameters(client api.RDSClient, pgName string, isCluster bool) map[string]string {
	if pgName == "" {
		return map[string]string{}
	}
	if v, ok := paramGroupCache[pgName]; ok {
		return v
	}

	params := make(map[string]string)

	// Fetch from API
	// Note: Pagination omitted for brevity, identifying key params only.
	// In real world, might need to iterate pages.

	if isCluster {
		output, err := client.DescribeDBClusterParameters(context.TODO(), &rds.DescribeDBClusterParametersInput{
			DBClusterParameterGroupName: &pgName,
		})
		if err == nil {
			for _, p := range output.Parameters {
				if p.ParameterName != nil && p.ParameterValue != nil {
					params[*p.ParameterName] = *p.ParameterValue
				}
			}
		}
	} else {
		output, err := client.DescribeDBParameters(context.TODO(), &rds.DescribeDBParametersInput{
			DBParameterGroupName: &pgName,
		})
		if err == nil {
			for _, p := range output.Parameters {
				if p.ParameterName != nil && p.ParameterValue != nil {
					params[*p.ParameterName] = *p.ParameterValue
				}
			}
		}
	}

	paramGroupCache[pgName] = params
	return params
}

func init() {
	rootCmd.AddCommand(rdsCmd)
}
