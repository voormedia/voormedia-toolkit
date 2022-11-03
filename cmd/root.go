package cmd

import (
	"github.com/spf13/cobra"
	"github.com/voormedia/voormedia-toolkit/pkg/util"
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
