package cmd

import (
	log "github.com/zakariakebairia/kvmcli/internal/logger"
	"github.com/zakariakebairia/kvmcli/internal/operations"
	"github.com/spf13/cobra"
)

// CreateCmd represents the command to create resource(s) from a manifest file.
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create resource(s) from a manifest file",
	Run: func(cmd *cobra.Command, args []string) {
		if ManifestPath == "" {
			log.Errorf("Manifest file is required (-f flag)")
		}

		// Use the provided configuration file to create resources.
		if err := operations.CreateFromManifest(ManifestPath); err != nil {
			log.Errorf("%v", err)
		}
	},
}

func init() {
	// Bind the manifest file flag to the global variable.
	CreateCmd.Flags().
		StringVarP(&ManifestPath, "file", "f", "", "Configuration file for the resource(s)")
}
