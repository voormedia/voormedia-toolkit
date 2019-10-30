package cmd

import (
	"github.com/emielvanlankveld/voormedia-toolkit/pkg/shell"
	"github.com/spf13/cobra"
)

var cmdShell = &cobra.Command{
	Use:   "shell",
	Short: "Run a shell of a Google Cloud SQL database.",
	DisableFlagsInUseLine: true,

	Run: func(_ *cobra.Command, args []string) {
		if err := shell.Run(log); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	cmdRoot.AddCommand(cmdShell)
}
