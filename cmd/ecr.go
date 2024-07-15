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
		levelInfo, levelWarning, _ := color.SetLevelColor()

		describeRepositories(client, tbl, levelWarning, levelInfo)

		table.Render("ECR", tbl)
	},
}

func init() {
	rootCmd.AddCommand(ecrCmd)
}

func describeRepositories(client *ecr.Client, table *tablewriter.Table, levelWarning string, levelInfo string) {
	resp, err := client.DescribeRepositories(context.TODO(), &ecr.DescribeRepositoriesInput{
		MaxResults: aws.Int32(100),
	})
	if err != nil {
		log.Fatalf("Failed to describe ECR repositories: %v", err)
	}

	for _, repo := range resp.Repositories {
		checkTagImmutability(repo, table, levelWarning)
		checkImageScanningConfiguration(repo, table, levelWarning)
		checkLifecyclePolicy(client, *repo.RepositoryName, table, levelInfo)
	}
}

func checkTagImmutability(repo types.Repository, table *tablewriter.Table, levelWarning string) {
	if repo.ImageTagMutability == types.ImageTagMutabilityMutable {
		table.Append([]string{"ECR", levelWarning, *repo.RepositoryName, "Tags can be overwritten"})
	}
}

func checkImageScanningConfiguration(repo types.Repository, table *tablewriter.Table, levelWarning string) {
	if !repo.ImageScanningConfiguration.ScanOnPush {
		table.Append([]string{"ECR", levelWarning, *repo.RepositoryName, "Image scanning is not enabled"})
	}
}

func checkLifecyclePolicy(client *ecr.Client, repoName string, table *tablewriter.Table, levelInfo string) {
	_, err := client.GetLifecyclePolicy(context.TODO(), &ecr.GetLifecyclePolicyInput{
		RepositoryName: aws.String(repoName),
	})
	if err != nil {
		var re *awshttp.ResponseError
		if errors.As(err, &re) && re.HTTPStatusCode() == 400 {
			table.Append([]string{"ECR", levelInfo, repoName, "Lifecycle policy is not set"})
		} else {
			log.Fatalf("Failed to describe lifecycle policy for repository %s: %v", repoName, err)
		}
	}
}
