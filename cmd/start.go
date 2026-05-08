package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zakariakebairia/kvmcli/internal/operations"
	"github.com/zakariakebairia/kvmcli/internal/providers/vm"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start resources such as virtual machines",
}

var startVMCmd = &cobra.Command{
	Use:   "vm <name>",
	Short: "Start a virtual machine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmName := args[0]

		session, cleanup, err := operations.NewSession(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}
		defer cleanup()

		if err := vm.Start(session, vmName); err != nil {
			return fmt.Errorf("failed to start vm %q: %w", vmName, err)
		}

		cmd.Printf("vm/%s started\n", vmName)
		return nil
	},
}

func init() {
	startCmd.AddCommand(startVMCmd)
}
