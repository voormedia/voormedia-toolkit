package cmd

import (
	"github.com/spf13/cobra"
	"github.com/voormedia/voormedia-toolkit/pkg/shell"
)

var cmdShell = &cobra.Command{
	Use:                   "shell",
	Short:                 "Run a shell of a Google Cloud SQL database.",
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
