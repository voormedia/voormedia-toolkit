package cmd

import (
	"github.com/emielvanlankveld/voormedia-toolkit/pkg/restore"
	"github.com/spf13/cobra"
	"os"
)

var cmdRestore = &cobra.Command{
	Use:   "restore",
	Short: "Download and restore a Backblaze backup of a Google Cloud SQL database.",
	DisableFlagsInUseLine: true,

	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		b2id, _ := cmd.Flags().GetString("b2id")
		b2key, _ := cmd.Flags().GetString("b2key")
		b2encrypt, _ := cmd.Flags().GetString("b2encrypt")
		b2bucket, _ := cmd.Flags().GetString("b2bucket")
		configFile, _ := cmd.Flags().GetString("dbconfig")
		targetPort, _ := cmd.Flags().GetString("port")
		targetHost, _ := cmd.Flags().GetString("host")
		targetUsername, _ := cmd.Flags().GetString("user")
		targetPassword, _ := cmd.Flags().GetString("password")
		targetDatabase, _ := cmd.Flags().GetString("database")
		if err := restore.Run(log, target, b2id, b2key, b2encrypt, b2bucket, configFile, targetPort, targetHost, targetUsername, targetPassword, targetDatabase); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	cmdRoot.AddCommand(cmdRestore)
	cmdRestore.Flags().String("target", "development", "Specifies the target environment of the backup restore")
	cmdRestore.Flags().String("b2id", os.Getenv("B2_ACCOUNT_ID"), "Specifies the Backblaze B2 account ID")
	cmdRestore.Flags().String("b2key", os.Getenv("B2_ACCOUNT_KEY"), "Specifies the Backblaze B2 account key")
	cmdRestore.Flags().String("b2encrypt", os.Getenv("B2_ENCRYPTION_KEY"), "Specifies the Backblaze B2 encryption key")
	cmdRestore.Flags().String("b2bucket", os.Getenv("B2_BACKUP_BUCKET"), "Specifies the Backblaze B2 backup bucket")
	cmdRestore.Flags().String("dbconfig", "./config/database.yml", "Specifies the location of the application's database configuration file")
	cmdRestore.Flags().String("port", "3306", "Specifies the port to use when restoring the target database")
	cmdRestore.Flags().String("host", "127.0.0.1", "Specifies the host to use when restoring the target database.")
	cmdRestore.Flags().String("user", "", "Specifies the user to use when restoring the target database. Only used when the database flag is also specified.")
	cmdRestore.Flags().String("password", "", "Specifies the password to use when restoring the target database. Only used when the database flag is also specified.")
	cmdRestore.Flags().String("database", "", "Specifies the name of the target database. Setting this option will override the configuration file.")
}
