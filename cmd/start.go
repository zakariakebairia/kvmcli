package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zakariakebairia/kvmcli/internal"
	"github.com/zakariakebairia/kvmcli/internal/providers/vm"
)

// CreateCmd represents the command to create resource(s) from a manifest file.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start resources like VMs",
}

var startVmCmd = &cobra.Command{
	Use:   "vm <vm-name>",
	Short: "Start a virtual machine",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		conn, err := internal.ConnectLibvirt()
		if err != nil {
			fmt.Println("init libvirt: %w", err)
		}
		// TODO: Add your VM starting logic here
		if err := vm.Start(conn, vmName); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("vm/%s started\n", vmName)
	},
}

func init() {
	// Bind the manifest file flag to the global variable.
	startCmd.AddCommand(startVmCmd)
}
