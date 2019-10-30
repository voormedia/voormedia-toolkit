package cmd

import (
	"github.com/emielvanlankveld/voormedia-toolkit/pkg/deploy"
	"github.com/spf13/cobra"
)

var cmdDeploy = &cobra.Command{
	Use:   "deploy",
	Short: "Builds, migrates and deployes the current application to GKE",

	Run: func(cmd *cobra.Command, args []string) {
		if err := deploy.Run(log); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	cmdRoot.AddCommand(cmdDeploy)
}
