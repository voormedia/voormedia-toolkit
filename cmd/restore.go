package cmd

import (
	"github.com/spf13/cobra"
	"github.com/voormedia/voormedia-toolkit/pkg/restore"
	"github.com/voormedia/voormedia-toolkit/pkg/util"
)

var cmdRestore = &cobra.Command{
	Use:                   "restore",
	Short:                 "Download and restore a Backblaze backup of a Google Cloud SQL database.",
	DisableFlagsInUseLine: true,

	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		shard, _ := cmd.Flags().GetString("shard")
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
		if err := restore.Run(log, target, shard, b2id, b2key, b2encrypt, b2bucket, configFile, targetPort, targetHost, targetUsername, targetPassword, targetDatabase); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	b2bucketName, b2encrypt, b2id, b2key := util.GetB2Config()

	cmdRoot.AddCommand(cmdRestore)
	cmdRestore.Flags().String("target", "development", "Specifies the target environment of the backup restore")
	cmdRestore.Flags().String("shard", "", "Specifies the shard that should be restored to (when multiple shards exist)")
	cmdRestore.Flags().String("b2id", b2id, "Specifies the Backblaze B2 account ID")
	cmdRestore.Flags().String("b2key", b2key, "Specifies the Backblaze B2 account key")
	cmdRestore.Flags().String("b2encrypt", b2encrypt, "Specifies the Backblaze B2 encryption key")
	cmdRestore.Flags().String("b2bucket", b2bucketName, "Specifies the Backblaze B2 backup bucket")
	cmdRestore.Flags().String("dbconfig", "./config/database.yml", "Specifies the location of the application's database configuration file")
	cmdRestore.Flags().String("port", "", "Specifies the port to use when restoring the target database")
	cmdRestore.Flags().String("host", "127.0.0.1", "Specifies the host to use when restoring the target database.")
	cmdRestore.Flags().String("user", "root", "Specifies the user to use when restoring the target database. Will be overriden by the database configuration file if applicable.")
	cmdRestore.Flags().String("password", "", "Specifies the password to use when restoring the target database. Only used when the database flag is also specified.")
	cmdRestore.Flags().String("database", "", "Specifies the name of the target database. Setting this option will override the configuration file.")
}
