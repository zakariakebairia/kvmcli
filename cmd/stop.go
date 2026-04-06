package cmd

import (
	"fmt"

	"github.com/zakariakebairia/kvmcli/internal"
	"github.com/zakariakebairia/kvmcli/internal/providers/vm"
	"github.com/spf13/cobra"
)

// CreateCmd represents the command to create resource(s) from a manifest file.
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop resources like VMs",
}

var stopVmCmd = &cobra.Command{
	Use:   "vm <vm-name>",
	Short: "Stop a virtual machine",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		conn, err := internal.InitConnection()
		if err != nil {
			fmt.Println("init libvirt: %w", err)
		}
		if err := vm.Stop(conn, vmName); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("vm/%s stopped\n", vmName)
	},
}

func init() {
	// Bind the manifest file flag to the global variable.
	stopCmd.AddCommand(stopVmCmd)
}
