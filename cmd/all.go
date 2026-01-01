package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Execute all commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Executing all subcommands...")
		albCmd.Run(albCmd, args)
		cloudfrontCmd.Run(cloudfrontCmd, args)
		cloudwatchlogsCmd.Run(cloudwatchlogsCmd, args)
		ecsCmd.Run(ecsCmd, args)
		ecrCmd.Run(ecrCmd, args)
		observabilityCmd.Run(observabilityCmd, args)
		rdsCmd.Run(rdsCmd, args)
		route53Cmd.Run(route53Cmd, args)
		s3Cmd.Run(s3Cmd, args)
		vpcCmd.Run(vpcCmd, args)
	},
}

func init() {
	rootCmd.AddCommand(allCmd)
}
