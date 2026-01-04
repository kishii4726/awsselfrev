package cmd

import (
	"context"
	"fmt"
	"log"

	"awsselfrev/internal/aws/api"
	"awsselfrev/internal/color"
	"awsselfrev/internal/config"
	"awsselfrev/internal/table"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var ecsCmd = &cobra.Command{
	Use:   "ecs",
	Short: "Check ECS configurations for best practices",
	Long: `This command checks various ECS configurations and best practices such as:
- Container Insights enabled
- Service circuit breaker (Warning)
- ARM64 architecture usage (Warning)`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		rules := config.LoadRules()
		tbl := table.SetTable()
		client := ecs.NewFromConfig(cfg)
		_, _, _ = color.SetLevelColor()

		checkECSConfigurations(client, tbl, rules)

		table.Render("ECS", tbl)
	},
}

func init() {
	rootCmd.AddCommand(ecsCmd)
}

func checkECSConfigurations(client api.ECSClient, tbl *tablewriter.Table, rules config.RulesConfig) {
	// 1. Check Clusters
	listResp, err := client.ListClusters(context.TODO(), &ecs.ListClustersInput{})
	if err != nil {
		log.Fatalf("Failed to list ECS clusters: %v", err)
	}

	if len(listResp.ClusterArns) == 0 {
		table.AddRow(tbl, []string{"ECS", "-", "-", "No clusters", "-", "-"})
		return
	}

	if len(listResp.ClusterArns) > 0 {
		descResp, err := client.DescribeClusters(context.TODO(), &ecs.DescribeClustersInput{
			Clusters: listResp.ClusterArns,
		})
		if err != nil {
			log.Fatalf("Failed to describe ECS clusters: %v", err)
		}

		for _, cluster := range descResp.Clusters {
			checkContainerInsights(cluster, tbl, rules)
			checkECSExecLogging(cluster, tbl, rules)
			checkServices(client, *cluster.ClusterArn, *cluster.ClusterName, tbl, rules)
		}
	}
}

func checkContainerInsights(cluster types.Cluster, tbl *tablewriter.Table, rules config.RulesConfig) {
	enabled := false
	for _, setting := range cluster.Settings {
		if setting.Name == types.ClusterSettingNameContainerInsights && setting.Value != nil && *setting.Value == "enabled" {
			enabled = true
			break
		}
	}

	rule := rules.Get("ecs-container-insights")
	if !enabled {
		table.AddRow(tbl, []string{rule.Service, "Fail", color.ColorizeLevel(rule.Level), *cluster.ClusterName, "Disabled", rule.Issue})
	} else {
		table.AddRow(tbl, []string{rule.Service, "Pass", "-", *cluster.ClusterName, "Enabled", rule.Issue})
	}
}

func checkECSExecLogging(cluster types.Cluster, tbl *tablewriter.Table, rules config.RulesConfig) {
	enabled := false
	if cluster.Configuration != nil && cluster.Configuration.ExecuteCommandConfiguration != nil {
		conf := cluster.Configuration.ExecuteCommandConfiguration
		if conf.Logging != types.ExecuteCommandLoggingNone {
			enabled = true
		}
	}

	rule := rules.Get("ecs-exec-logging")
	if !enabled {
		table.AddRow(tbl, []string{rule.Service, "Fail", color.ColorizeLevel(rule.Level), *cluster.ClusterName, "Disabled", rule.Issue})
	} else {
		table.AddRow(tbl, []string{rule.Service, "Pass", "-", *cluster.ClusterName, "Enabled", rule.Issue})
	}
}

func checkServices(client api.ECSClient, clusterArn string, clusterName string, tbl *tablewriter.Table, rules config.RulesConfig) {
	// List Services
	// Note: Pagination should be handled for production, but kept simple for now as per previous pattern.
	svcResp, err := client.ListServices(context.TODO(), &ecs.ListServicesInput{
		Cluster: &clusterArn,
	})
	if err != nil {
		log.Fatalf("Failed to list services for cluster %s: %v", clusterName, err)
	}

	if len(svcResp.ServiceArns) > 0 {
		descResp, err := client.DescribeServices(context.TODO(), &ecs.DescribeServicesInput{
			Cluster:  &clusterArn,
			Services: svcResp.ServiceArns,
		})
		if err != nil {
			log.Fatalf("Failed to describe services for cluster %s: %v", clusterName, err)
		}

		for _, service := range descResp.Services {
			checkCircuitBreaker(service, tbl, rules)
			checkCpuArchitectureAndSensitiveInfo(client, service, tbl, rules)
			checkPropagateTags(service, tbl, rules)
		}
	}
}

func checkPropagateTags(service types.Service, tbl *tablewriter.Table, rules config.RulesConfig) {
	rule := rules.Get("ecs-propagate-tags")
	if service.PropagateTags == types.PropagateTagsNone {
		table.AddRow(tbl, []string{rule.Service, "Fail", color.ColorizeLevel(rule.Level), *service.ServiceName, string(service.PropagateTags), rule.Issue})
	} else {
		table.AddRow(tbl, []string{rule.Service, "Pass", "-", *service.ServiceName, string(service.PropagateTags), rule.Issue})
	}
}

func checkCircuitBreaker(service types.Service, tbl *tablewriter.Table, rules config.RulesConfig) {
	// Circuit breaker is in DeploymentConfiguration
	enabled := false
	if service.DeploymentConfiguration != nil &&
		service.DeploymentConfiguration.DeploymentCircuitBreaker != nil &&
		service.DeploymentConfiguration.DeploymentCircuitBreaker.Enable {
		enabled = true
	}

	rule := rules.Get("ecs-service-circuit-breaker")
	if !enabled {
		table.AddRow(tbl, []string{rule.Service, "Fail", color.ColorizeLevel(rule.Level), *service.ServiceName, "Disabled", rule.Issue})
	} else {
		table.AddRow(tbl, []string{rule.Service, "Pass", "-", *service.ServiceName, "Enabled", rule.Issue})
	}
}

func checkCpuArchitectureAndSensitiveInfo(client api.ECSClient, service types.Service, tbl *tablewriter.Table, rules config.RulesConfig) {
	// We need to look at the Task Definition
	// service.TaskDefinition is an ARN.
	if service.TaskDefinition == nil {
		return
	}

	tdResp, err := client.DescribeTaskDefinition(context.TODO(), &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: service.TaskDefinition,
	})
	if err != nil {
		log.Printf("Failed to describe task definition %s: %v", *service.TaskDefinition, err)
		return
	}

	// 1. Check CPU Architecture
	// CPU Architecture is in RuntimePlatform
	isArm64 := false
	if tdResp.TaskDefinition.RuntimePlatform != nil && tdResp.TaskDefinition.RuntimePlatform.CpuArchitecture == types.CPUArchitectureArm64 {
		isArm64 = true
	}

	arch := "Unknown"
	if tdResp.TaskDefinition.RuntimePlatform != nil {
		arch = string(tdResp.TaskDefinition.RuntimePlatform.CpuArchitecture)
	}

	ruleArch := rules.Get("ecs-cpu-architecture")
	if !isArm64 {
		table.AddRow(tbl, []string{ruleArch.Service, "Fail", color.ColorizeLevel(ruleArch.Level), *service.ServiceName, arch, ruleArch.Issue})
	} else {
		table.AddRow(tbl, []string{ruleArch.Service, "Pass", "-", *service.ServiceName, arch, ruleArch.Issue})
	}

	// 2. Check Sensitive Environment Variables
	checkSensitiveEnvironmentVariables(tdResp.TaskDefinition, service.ServiceName, tbl, rules)
}

func checkSensitiveEnvironmentVariables(td *types.TaskDefinition, serviceName *string, tbl *tablewriter.Table, rules config.RulesConfig) {
	sensitiveKeywords := []string{"PASSWORD", "TOKEN", "SECRET", "KEY", "CREDENTIAL"}
	foundSensitive := false
	var foundKeys []string

	for _, container := range td.ContainerDefinitions {
		for _, env := range container.Environment {
			upperKey := strings.ToUpper(*env.Name)
			for _, keyword := range sensitiveKeywords {
				if strings.Contains(upperKey, keyword) {
					foundSensitive = true
					foundKeys = append(foundKeys, *env.Name)
					break
				}
			}
		}
	}

	rule := rules.Get("ecs-sensitive-environment-variables")
	resourceName := *serviceName
	if foundSensitive {
		status := fmt.Sprintf("Found: %s", strings.Join(foundKeys, ", "))
		table.AddRow(tbl, []string{rule.Service, "Fail", color.ColorizeLevel(rule.Level), resourceName, status, rule.Issue})
	} else {
		table.AddRow(tbl, []string{rule.Service, "Pass", "-", resourceName, "Safe", rule.Issue})
	}
}
