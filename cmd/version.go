package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/akerl/ledgersql/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of ledgersql",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s\n", version.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
