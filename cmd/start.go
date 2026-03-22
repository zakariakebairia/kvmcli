package cmd

import (
	"context"
	"fmt"

	"github.com/zakariakebairia/kvmcli/internal"
	"github.com/zakariakebairia/kvmcli/internal/vms"
	"github.com/spf13/cobra"
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
		conn, err := internal.InitConnection()
		if err != nil {
			fmt.Println("init libvirt: %w", err)
		}
		// TODO: Add your VM starting logic here
		ctx := context.Background()
		dom := vms.NewLibvirtDomainManager(conn)
		dom.Start(ctx, vmName)
		fmt.Printf("vm/%s started\n", vmName)
	},
}

func init() {
	// Bind the manifest file flag to the global variable.
	startCmd.AddCommand(startVmCmd)
}
