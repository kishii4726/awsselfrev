package cmd

import (
	"context"
	"errors"
	"log"

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

		describeRepositories(client, tbl, rules)

		table.Render("ECR", tbl)
	},
}

func init() {
	rootCmd.AddCommand(ecrCmd)
}

func describeRepositories(client *ecr.Client, table *tablewriter.Table, rules config.RulesConfig) {
	resp, err := client.DescribeRepositories(context.TODO(), &ecr.DescribeRepositoriesInput{
		MaxResults: aws.Int32(100),
	})
	if err != nil {
		log.Fatalf("Failed to describe ECR repositories: %v", err)
	}

	for _, repo := range resp.Repositories {
		checkTagImmutability(repo, table, rules)
		checkImageScanningConfiguration(repo, table, rules)
		checkLifecyclePolicy(client, *repo.RepositoryName, table, rules)
	}
}

func checkTagImmutability(repo types.Repository, table *tablewriter.Table, rules config.RulesConfig) {
	if repo.ImageTagMutability == types.ImageTagMutabilityMutable {
		rule := rules.Rules["ecr-tag-immutability"]
		table.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), *repo.RepositoryName, rule.Issue})
	}
}

func checkImageScanningConfiguration(repo types.Repository, table *tablewriter.Table, rules config.RulesConfig) {
	if !repo.ImageScanningConfiguration.ScanOnPush {
		rule := rules.Rules["ecr-image-scanning"]
		table.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), *repo.RepositoryName, rule.Issue})
	}
}

func checkLifecyclePolicy(client *ecr.Client, repoName string, table *tablewriter.Table, rules config.RulesConfig) {
	_, err := client.GetLifecyclePolicy(context.TODO(), &ecr.GetLifecyclePolicyInput{
		RepositoryName: aws.String(repoName),
	})
	if err != nil {
		var re *awshttp.ResponseError
		if errors.As(err, &re) && re.HTTPStatusCode() == 400 {
			rule := rules.Rules["ecr-lifecycle-policy"]
			table.Append([]string{rule.Service, color.ColorizeLevel(rule.Level), repoName, rule.Issue})
		} else {
			log.Fatalf("Failed to describe lifecycle policy for repository %s: %v", repoName, err)
		}
	}
}
