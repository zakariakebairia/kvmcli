package cmd

import (
	"fmt"

	log "github.com/zakariakebairia/kvmcli/internal/logger"
	"github.com/zakariakebairia/kvmcli/internal/operations"
	"github.com/spf13/cobra"
)

// Create the "get" parent command.
var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve information about VMs, snapshots, networks, etc.",
}

// 'get vm' subcommand: shows virtual machines.
var GetVMCmd = &cobra.Command{
	Use:   "vm",
	Short: "Display information about virtual machines",
	Run: func(cmd *cobra.Command, args []string) {
		if Namespace != "" {
			// operations.ListByNamespace(Namespace)
			fmt.Println("List by namespace here for vms")
			// operations.ListAll()
			return
		}
		operations.ListAll()
	},
}

// 'get snapshots' subcommand: shows snapshots.
var GetSnapshotsCmd = &cobra.Command{
	Use:     "snapshot",
	Aliases: []string{"snap"},
	Short:   "Display snapshots for virtual machines",
	Run: func(cmd *cobra.Command, args []string) {
		if ConfigFile == "" {
			log.Errorf("Configuration file is required (-f flag)")
		}
		fmt.Println("You snapshosts are here")
		// op.ListSnapshost()
	},
}

// 'get snapshots' subcommand: shows snapshots.
var GetNetworkCmd = &cobra.Command{
	Use:     "network",
	Aliases: []string{"net"},
	Short:   "Display network details",
	Run: func(cmd *cobra.Command, args []string) {
		if Namespace != "" {
			// network.ListByNamespace(Namespace)
			fmt.Println("List by namespace here for networks")
			return
		}
		operations.ListAllNetworks()
	},
}

// 'get store' subcommand: shows stores.
var GetStoreCmd = &cobra.Command{
	Use:     "store",
	Aliases: []string{"st"},
	Short:   "Display stores details",
	Run: func(cmd *cobra.Command, args []string) {
		if Namespace != "" {
			// TODO: implement ListStoresByNamespace
			fmt.Println("List by namespace here for stores")
			return
		}
		operations.ListAllStores()
	},
}

func init() {
	// Flags for virtual machines
	GetVMCmd.Flags().
		StringVarP(&Namespace, "namespace", "n", "", "Namespace")
		// Flags for Networks
	GetNetworkCmd.Flags().
		StringVarP(&Namespace, "namespace", "n", "", "Namespace")
		// Flags for Snapshots
	GetSnapshotsCmd.Flags().
		StringVarP(&Namespace, "namespace", "n", "", "Namespace")
		// Flags for stores
	GetStoreCmd.Flags().
		StringVarP(&Namespace, "namespace", "n", "", "Namespace")
	GetCmd.AddCommand(GetVMCmd, GetSnapshotsCmd, GetNetworkCmd, GetStoreCmd)
}
