package cmd

import (
	"github.com/spf13/cobra"
	"github.com/voormedia/voormedia-toolkit/pkg/backup"
	"github.com/voormedia/voormedia-toolkit/pkg/util"
)

var cmdBackup = &cobra.Command{
	Use:   "backup",
	Short: "Creates a new backup of a Google Cloud SQL database and stores it in Backblaze B2.",

	Run: func(cmd *cobra.Command, args []string) {
		shard, _ := cmd.Flags().GetString("shard")
		b2id, _ := cmd.Flags().GetString("b2id")
		b2key, _ := cmd.Flags().GetString("b2key")
		b2encrypt, _ := cmd.Flags().GetString("b2encrypt")
		b2bucket, _ := cmd.Flags().GetString("b2bucket")
		port, _ := cmd.Flags().GetString("port")
		host, _ := cmd.Flags().GetString("host")
		configFile, _ := cmd.Flags().GetString("dbconfig")
		if err := backup.Run(log, port, host, shard, b2id, b2key, b2encrypt, b2bucket, configFile); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	b2bucketName, b2encrypt, b2id, b2key := util.GetB2Config()

	cmdRoot.AddCommand(cmdBackup)
	cmdBackup.Flags().String("shard", "", "Specifies the shard that should be backup up (when multiple shards exist)")
	cmdBackup.Flags().String("b2id", b2id, "Specifies the Backblaze B2 account ID")
	cmdBackup.Flags().String("b2key", b2key, "Specifies the Backblaze B2 account key")
	cmdBackup.Flags().String("b2encrypt", b2encrypt, "Specifies the Backblaze B2 encryption key")
	cmdBackup.Flags().String("b2bucket", b2bucketName, "Specifies the Backblaze B2 backup bucket")
	cmdBackup.Flags().String("port", "3307", "Specifies the port to use to reach the source database")
	cmdBackup.Flags().String("host", "127.0.0.1", "Specifies the host to use to reach the source database.")
	cmdBackup.Flags().String("dbconfig", "./config/database.yml", "Specifies the location of the application's database configuration file")
}
