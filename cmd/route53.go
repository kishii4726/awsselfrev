package cmd

import (
	"context"
	"log"

	"awsselfrev/internal/aws/api"
	"awsselfrev/internal/color"
	"awsselfrev/internal/config"
	"awsselfrev/internal/table"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var route53Cmd = &cobra.Command{
	Use:   "route53",
	Short: "Check Route53 configurations for best practices",
	Long: `This command checks various Route53 configurations and best practices such as:
- Query logging enabled`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		rules := config.LoadRules()
		tbl := table.SetTable()
		client := route53.NewFromConfig(cfg)
		_, _, _ = color.SetLevelColor()

		checkRoute53Configurations(client, tbl, rules)

		table.Render("Route53", tbl)
	},
}

func init() {
	rootCmd.AddCommand(route53Cmd)
}

func checkRoute53Configurations(client api.Route53Client, tbl *tablewriter.Table, rules config.RulesConfig) {
	// List Hosted Zones
	zones, err := client.ListHostedZones(context.TODO(), &route53.ListHostedZonesInput{})
	if err != nil {
		log.Fatalf("Failed to list hosted zones: %v", err)
	}

	if len(zones.HostedZones) == 0 {
		table.AddRow(tbl, []string{"Route53", "-", "-", "No hosted zones", "-", "-"})
		return
	}

	for _, zone := range zones.HostedZones {
		// Public zones do not necessarily need query logging, but the requirement was "Route53 Query Logs enabled"
		// Typically this applies to public zones or useful for auditing. The requirement didn't specify public/private.
		// We will check if query logging config exists for the zone.
		if zone.Config.PrivateZone {
			// Skip private zones? Usually query logging is for public, but can be for private resolver logs.
			// Let's assume we check for all or follow typical best practices. AWS Config rule checks if query logging is enabled.
			// We will check all.
		}

		configs, err := client.ListQueryLoggingConfigs(context.TODO(), &route53.ListQueryLoggingConfigsInput{
			HostedZoneId: zone.Id,
		})
		if err != nil {
			log.Fatalf("Failed to list query logging configs for zone %s: %v", *zone.Id, err)
		}

		rule := rules.Get("route53-query-logging")
		if len(configs.QueryLoggingConfigs) == 0 {
			table.AddRow(tbl, []string{rule.Service, "Fail", color.ColorizeLevel(rule.Level), *zone.Name, "Disabled", rule.Issue})
		} else {
			table.AddRow(tbl, []string{rule.Service, "Pass", "-", *zone.Name, "Enabled", rule.Issue})
		}
	}
}
