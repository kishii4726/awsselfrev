package cmd

import (
	"context"
	"errors"
	"log"
	"strings"

	"awsselfrev/internal/aws/api"
	"awsselfrev/internal/color"
	"awsselfrev/internal/config"
	"awsselfrev/internal/table"

	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin/types"
	"github.com/aws/smithy-go"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var observabilityCmd = &cobra.Command{
	Use:   "observability",
	Short: "Check Observability configurations for best practices",
	Long: `This command checks various Observability configurations and best practices such as:
- Telemetry resource tags enablement`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		rules := config.LoadRules()
		tbl := table.SetTable()
		client := observabilityadmin.NewFromConfig(cfg)
		_, _, _ = color.SetLevelColor()

		checkTelemetryResourceTags(client, tbl, rules)

		table.Render("Observability", tbl)
	},
}

func init() {
	rootCmd.AddCommand(observabilityCmd)
}

func checkTelemetryResourceTags(client api.ObservabilityAdminClient, table *tablewriter.Table, rules config.RulesConfig) {
	resp, err := client.GetTelemetryEnrichmentStatus(context.TODO(), &observabilityadmin.GetTelemetryEnrichmentStatusInput{})
	rule := rules.Get("telemetry-resource-tags-enabled")
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) && strings.Contains(ae.ErrorCode(), "ResourceNotFoundException") {
			// If not found, it means it's not enabled.
			table.Append([]string{rule.Service, "Fail", color.ColorizeLevel(rule.Level), "Account", rule.Issue})
			return
		}
		log.Printf("Failed to get telemetry enrichment status: %v", err)
		return
	}

	if resp.Status != types.TelemetryEnrichmentStatusRunning {
		table.Append([]string{rule.Service, "Fail", color.ColorizeLevel(rule.Level), "Account", rule.Issue})
	} else {
		table.Append([]string{rule.Service, "Pass", "-", "Account", rule.Issue})
	}
}
