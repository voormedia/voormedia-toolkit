package cmd

import (
	"github.com/emielvanlankveld/gsql/pkg/util"
	"github.com/spf13/cobra"
)

var version = "master"
var log = util.NewLogger("gsql")

var verbose bool

var cmdRoot = &cobra.Command{
	Use:   "gsql",
	Short: "Perform actions on Google Cloud SQL databases.",
}

func init() {
	cmdRoot.Version = version
	cmdRoot.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func Execute() {
	cmdRoot.Execute()
}
