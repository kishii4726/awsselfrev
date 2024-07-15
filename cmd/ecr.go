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
		tbl := table.SetTable()
		client := ecr.NewFromConfig(cfg)
		level_info, level_warning, _ := color.SetLevelColor()

		describeRepositories(client, tbl, level_warning, level_info)

		table.Render("ECR", tbl)
	},
}

func init() {
	rootCmd.AddCommand(ecrCmd)
}

func describeRepositories(client *ecr.Client, table *tablewriter.Table, level_warning, level_info string) {
	resp, err := client.DescribeRepositories(context.TODO(), &ecr.DescribeRepositoriesInput{
		MaxResults: aws.Int32(100),
	})
	if err != nil {
		log.Fatalf("Failed to describe ECR repositories: %v", err)
	}

	for _, repo := range resp.Repositories {
		checkTagImmutability(repo, table, level_warning)
		checkImageScanningConfiguration(repo, table, level_warning)
		checkLifecyclePolicy(client, *repo.RepositoryName, table, level_info)
	}
}

func checkTagImmutability(repo types.Repository, table *tablewriter.Table, level_warning string) {
	if repo.ImageTagMutability == types.ImageTagMutabilityMutable {
		table.Append([]string{"ECR", level_warning, *repo.RepositoryName, "Tags can be overwritten"})
	}
}

func checkImageScanningConfiguration(repo types.Repository, table *tablewriter.Table, level_warning string) {
	if !repo.ImageScanningConfiguration.ScanOnPush {
		table.Append([]string{"ECR", level_warning, *repo.RepositoryName, "Image scanning is not enabled"})
	}
}

func checkLifecyclePolicy(client *ecr.Client, repoName string, table *tablewriter.Table, level_info string) {
	_, err := client.GetLifecyclePolicy(context.TODO(), &ecr.GetLifecyclePolicyInput{
		RepositoryName: aws.String(repoName),
	})
	if err != nil {
		var re *awshttp.ResponseError
		if errors.As(err, &re) && re.HTTPStatusCode() == 400 {
			table.Append([]string{"ECR", level_info, repoName, "Lifecycle policy is not set"})
		} else {
			log.Fatalf("Failed to describe lifecycle policy for repository %s: %v", repoName, err)
		}
	}
}
