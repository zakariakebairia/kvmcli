package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "dev"

var ShowVersion = &cobra.Command{
	Use:   "version",
	Short: "Show current version of kvmcli",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}
