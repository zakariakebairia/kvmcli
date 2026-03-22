package cmd

import (
	log "github.com/zakariakebairia/kvmcli/internal/logger"
	"github.com/zakariakebairia/kvmcli/internal/operations"
	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete resource(s) from a manifest file",
	Run: func(cmd *cobra.Command, args []string) {
		if DeleteAll {
			// Delete all VMs
			return
		} else if ManifestPath == "" {
			log.Errorf("Manifest file is required (-f flag)")
		}
		// Call your delete operation with the provided file.
		operations.DeleteFromManifest(ManifestPath)
	},
}

func init() {
	DeleteCmd.Flags().
		StringVarP(&ManifestPath, "file", "f", "", "Manifest file for the resource(s) to delete")
	// DeleteCmd.Flags().BoolVar(&DeleteAll, "all", false, "Delete all VMs")
}
