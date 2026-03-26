package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ShowVersion = &cobra.Command{
	Use:   "version",
	Short: "Show current version of kvmcli",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("v1.0.0-alpha")
	},
}
