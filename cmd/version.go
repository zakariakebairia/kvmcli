package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "unknown"
	Built   = "unknown"
)

var ShowVersion = &cobra.Command{
	Use:   "version",
	Short: "Show current version of kvmcli",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("kvmcli %s\n", Version)
		fmt.Printf("  commit:  %s\n", Commit)
		fmt.Printf("  built:   %s\n", Built)
		fmt.Printf("  go:      %s\n", runtime.Version())
	},
}
