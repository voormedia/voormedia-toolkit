package cmd

import (
	"github.com/emielvanlankveld/voormedia-toolkit/pkg/util"
	"github.com/spf13/cobra"
)

var version = "master"
var log = util.NewLogger("voormedia-toolkit")

var verbose bool

var cmdRoot = &cobra.Command{
	Use:   "vmt",
	Short: "Useful utilities related to apps deployed on Google Cloud.",
}

func init() {
	cmdRoot.Version = version
	cmdRoot.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func Execute() {
	cmdRoot.Execute()
}
