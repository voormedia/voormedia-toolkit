package cmd

import (
	"github.com/emielvanlankveld/voormedia-toolkit/pkg/backup"
	"github.com/spf13/cobra"
)

var cmdBackup = &cobra.Command{
	Use:   "backup",
	Short: "Creates a new backup of a Google Cloud SQL database and stores it in Backblaze B2.",

	Run: func(cmd *cobra.Command, args []string) {
		if err := backup.Run(log); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	cmdRoot.AddCommand(cmdBackup)
}
