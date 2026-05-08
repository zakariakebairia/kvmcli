package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zakariakebairia/kvmcli/internal/operations"
	"github.com/zakariakebairia/kvmcli/internal/providers/vm"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		vmName := args[0]
		// fmt.Println("%s, %s", args[0], args[1])

		session, cleanup, err := operations.NewSession(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}
		defer cleanup()

		if err := vm.Stop(session, vmName); err != nil {
			return fmt.Errorf("failed to stop vm %q: %w", vmName, err)
		}

		cmd.Printf("vm/%s stopped\n", vmName)
		return nil
	},
}

func init() {
	// Bind the manifest file flag to the global variable.
	stopCmd.AddCommand(stopVmCmd)
}
