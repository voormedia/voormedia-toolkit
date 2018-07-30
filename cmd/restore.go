package cmd

import (
	"github.com/emielvanlankveld/gsql/pkg/restore"
	"github.com/spf13/cobra"
)

var cmdRestore = &cobra.Command{
	Use:   "restore",
	Short: "Download and restore a copy of a Google Cloud SQL database.",
	DisableFlagsInUseLine: true,

	Run: func(_ *cobra.Command, args []string) {
		if err := restore.Run(log); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	cmdRoot.AddCommand(cmdRestore)
}
