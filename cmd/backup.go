package cmd

import (
	"github.com/emielvanlankveld/voormedia-toolkit/pkg/backup"
	"github.com/spf13/cobra"
	"os"
)

var cmdBackup = &cobra.Command{
	Use:   "backup",
	Short: "Creates a new backup of a Google Cloud SQL database and stores it in Backblaze B2.",

	Run: func(cmd *cobra.Command, args []string) {
		b2id, _ := cmd.Flags().GetString("b2id")
		b2key, _ := cmd.Flags().GetString("b2key")
		b2encrypt, _ := cmd.Flags().GetString("b2encrypt")
		b2bucket, _ := cmd.Flags().GetString("b2bucket")
		port, _ := cmd.Flags().GetString("port")
		host, _ := cmd.Flags().GetString("host")
		configFile, _ := cmd.Flags().GetString("dbconfig")
		if err := backup.Run(log, port, host, b2id, b2key, b2encrypt, b2bucket, configFile); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	cmdRoot.AddCommand(cmdBackup)
	cmdBackup.Flags().String("b2id", os.Getenv("B2_ACCOUNT_ID"), "Specifies the Backblaze B2 account ID")
	cmdBackup.Flags().String("b2key", os.Getenv("B2_ACCOUNT_KEY"), "Specifies the Backblaze B2 account key")
	cmdBackup.Flags().String("b2encrypt", os.Getenv("B2_ENCRYPTION_KEY"), "Specifies the Backblaze B2 encryption key")
	cmdBackup.Flags().String("b2bucket", "voormedia-db-backups", "Specifies the Backblaze B2 backup bucket")
	cmdBackup.Flags().String("port", "3307", "Specifies the port to use to reach the source database")
	cmdBackup.Flags().String("host", "127.0.0.1", "Specifies the host to use to reach the source database.")
	cmdBackup.Flags().String("dbconfig", "./config/database.yml", "Specifies the location of the application's database configuration file")
}
