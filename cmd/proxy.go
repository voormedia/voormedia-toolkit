package cmd

import (
	"github.com/spf13/cobra"
	"github.com/voormedia/voormedia-toolkit/pkg/proxy"
)

var cmdProxy = &cobra.Command{
	Use:   "proxy",
	Short: "Connect to Google Cloud SQL instance through a proxy.",

	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetString("port")
		if err := proxy.Run(log, port); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	cmdRoot.AddCommand(cmdProxy)
	cmdProxy.Flags().String("port", "3307", "Specifies the port the proxy should connect on")
}
