package cmd

import (
	"text/template"

	"github.com/spf13/cobra"
)

var cmdVersion = &cobra.Command{
	Use:                   "version",
	Short:                 "Show version",
	DisableFlagsInUseLine: true,

	Long: "Show version number and exit",

	Run: func(cmd *cobra.Command, args []string) {
		tmpl := template.Must(template.New("top").Parse(cmd.VersionTemplate()))
		tmpl.Execute(cmd.OutOrStdout(), cmdRoot)
	},
}

func init() {
	cmdRoot.AddCommand(cmdVersion)
}
