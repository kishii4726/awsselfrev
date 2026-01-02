package cmd

import (
	"context"
	"errors"
	"log"

	"awsselfrev/internal/aws/api"
	"awsselfrev/internal/color"
	"awsselfrev/internal/config"
	"awsselfrev/internal/table"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var ecrCmd = &cobra.Command{
	Use:   "ecr",
	Short: "Checks ECR configurations for best practices",
	Long: `This command checks various ECR configurations and best practices such as:
- Tag immutability
- Image scanning configuration
- Lifecycle policy`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		rules := config.LoadRules()
		tbl := table.SetTable()
		client := ecr.NewFromConfig(cfg)
		_, _, _ = color.SetLevelColor()

		checkECRConfigurations(client, tbl, rules)

		table.Render("ECR", tbl)
	},
}

func init() {
	rootCmd.AddCommand(ecrCmd)
}

func checkECRConfigurations(client api.ECRClient, tbl *tablewriter.Table, rules config.RulesConfig) {
	resp, err := client.DescribeRepositories(context.TODO(), &ecr.DescribeRepositoriesInput{
		MaxResults: aws.Int32(100),
	})
	if err != nil {
		log.Fatalf("Failed to describe ECR repositories: %v", err)
	}

	if len(resp.Repositories) == 0 {
		table.AddRow(tbl, []string{"ECR", "-", "-", "No repositories", "-", "-"})
		return
	}

	for _, repo := range resp.Repositories {
		checkTagImmutability(repo, tbl, rules)
		checkImageScanningConfiguration(repo, tbl, rules)
		checkLifecyclePolicy(client, *repo.RepositoryName, tbl, rules)
	}
}

func checkTagImmutability(repo types.Repository, tbl *tablewriter.Table, rules config.RulesConfig) {
	rule := rules.Get("ecr-tag-immutability")
	if repo.ImageTagMutability == types.ImageTagMutabilityMutable {
		table.AddRow(tbl, []string{rule.Service, "Fail", color.ColorizeLevel(rule.Level), *repo.RepositoryName, "Mutable", rule.Issue})
	} else {
		table.AddRow(tbl, []string{rule.Service, "Pass", "-", *repo.RepositoryName, "Immutable", rule.Issue})
	}
}

func checkImageScanningConfiguration(repo types.Repository, tbl *tablewriter.Table, rules config.RulesConfig) {
	rule := rules.Get("ecr-image-scanning")
	if !repo.ImageScanningConfiguration.ScanOnPush {
		table.AddRow(tbl, []string{rule.Service, "Fail", color.ColorizeLevel(rule.Level), *repo.RepositoryName, "Disabled", rule.Issue})
	} else {
		table.AddRow(tbl, []string{rule.Service, "Pass", "-", *repo.RepositoryName, "Enabled", rule.Issue})
	}
}

func checkLifecyclePolicy(client api.ECRClient, repoName string, tbl *tablewriter.Table, rules config.RulesConfig) {
	_, err := client.GetLifecyclePolicy(context.TODO(), &ecr.GetLifecyclePolicyInput{
		RepositoryName: aws.String(repoName),
	})
	rule := rules.Get("ecr-lifecycle-policy")
	if err != nil {
		var re *awshttp.ResponseError
		if errors.As(err, &re) && re.HTTPStatusCode() == 400 {
			table.AddRow(tbl, []string{rule.Service, "Fail", color.ColorizeLevel(rule.Level), repoName, "Missing", rule.Issue})
		} else {
			log.Fatalf("Failed to describe lifecycle policy for repository %s: %v", repoName, err)
		}
	} else {
		table.AddRow(tbl, []string{rule.Service, "Pass", "-", repoName, "Set", rule.Issue})
	}
}
