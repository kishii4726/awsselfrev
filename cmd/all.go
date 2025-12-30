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
		ec2Cmd.Run(ec2Cmd, args)
		vpcCmd.Run(vpcCmd, args)
		s3Cmd.Run(s3Cmd, args)
		rdsCmd.Run(rdsCmd, args)
		ecrCmd.Run(ecrCmd, args)
		cloudwatchlogsCmd.Run(cloudwatchlogsCmd, args)
		cloudfrontCmd.Run(cloudfrontCmd, args)
		observabilityCmd.Run(observabilityCmd, args)
	},
}

func init() {
	rootCmd.AddCommand(allCmd)
}
