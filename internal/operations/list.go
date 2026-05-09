package operations

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/zakariakebairia/kvmcli/internal/database"
	"github.com/zakariakebairia/kvmcli/internal/registry"

	"github.com/zakariakebairia/kvmcli/internal/providers/vm"
)

// ListAll retrieves and prints all resources of the given kind in a tabular format.
func ListAll(ctx context.Context, kind string) error {
	descriptor, ok := registry.Get(kind)
	if !ok {
		return fmt.Errorf("unknown resource kind: %q", kind)
	}

	session, cleanup, err := NewSession(ctx)
	if err != nil {
		return fmt.Errorf("creating session: %w", err)
	}
	defer cleanup()

	dbHandler := database.NewDBHandler(session.DB)
	if err := dbHandler.EnsureTable(ctx); err != nil {
		return fmt.Errorf("ensuring table for %q: %w", kind, err)
	}

	objects, err := dbHandler.List(ctx, kind)
	if err != nil {
		return fmt.Errorf("listing %q resources: %w", kind, err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, strings.Join(descriptor.Columns, "\t"))
	for _, object := range objects {

		// vm exception
		if kind == "vm" {
			status, err := vm.GetVMStatus(session, object.Name)
			if err != nil {
				return fmt.Errorf("get vm status: %w", err)
			}

			object.Status = status
		}
		//
		fmt.Fprintln(w, strings.Join(descriptor.Format(object), "\t"))
	}

	return w.Flush()
}
