package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version shifuctl",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("shifuctl v0.3.1")
	},
}
