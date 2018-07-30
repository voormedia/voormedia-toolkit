package cmd

import (
	"github.com/emielvanlankveld/gsql/pkg/proxy"
	"github.com/spf13/cobra"
)

var cmdProxy = &cobra.Command{
	Use:   "proxy",
	Short: "Connect to Google Cloud SQL instance through a proxy.",
	DisableFlagsInUseLine: true,

	Run: func(_ *cobra.Command, args []string) {
		if err := proxy.Run(log); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	cmdRoot.AddCommand(cmdProxy)
}
