package operations

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/zakariakebairia/kvmcli/internal/database"
	"github.com/zakariakebairia/kvmcli/internal/registry"
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

// IDEA: resources, err := operator.GetResources()
//
//		if err != nil {
//			return err
//	}
//
// w := resource.Header()
//
//	for _, resource := range resources {
//		resource.PrintRow(w)
//	}
//
// -------------------------------------------
//
//	func ListAll() error {
//		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//		defer cancel()
//
//		operator, err := NewOperator(ctx, "")
//		if err != nil {
//			return fmt.Errorf("failed to initialize operator: %w", err)
//		}
//		defer operator.Close()
//
//		virtualMachines, err := vms.GetVirtualMachines(operator.ctx, operator.db, operator.conn)
//		if err != nil {
//			return fmt.Errorf("failed to retrieve VMs: %w", err)
//		}
//
//		info := &vms.VirtualMachineInfo{}
//		w := info.Header()
//
//		for _, virtualMachine := range virtualMachines {
//			virtualMachine.PrintInfo(w)
//		}
//		w.Flush()
//		return nil
//	}
//
//	func ListAllNetworks() error {
//		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//		defer cancel()
//
//		operator, err := NewOperator(ctx, "")
//		if err != nil {
//			return fmt.Errorf("failed to initialize operator: %w", err)
//		}
//		defer operator.Close()
//
//		virtualNetworks, err := network.GetVirtualNetworks(operator.ctx, operator.db, operator.conn)
//		if err != nil {
//			return fmt.Errorf("failed to retrieve networks: %w", err)
//		}
//
//		info := &network.VirtualNetworkInfo{}
//		w := info.Header()
//
//		for _, virtualNetwork := range virtualNetworks {
//			virtualNetwork.PrintInfo(w)
//		}
//		w.Flush()
//		return nil
//	}
//
// // ListAllStores lists all stores in the database.
//
//	func ListAllStores() error {
//		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//		defer cancel()
//
//		operator, err := NewOperator(ctx, "")
//		if err != nil {
//			return fmt.Errorf("failed to initialize operator: %w", err)
//		}
//		defer operator.Close()
//
//		stores, err := store.GetStores(operator.ctx, operator.db)
//		if err != nil {
//			return fmt.Errorf("failed to retrieve stores: %w", err)
//		}
//
//		info := &store.StoreInfo{}
//		w := info.Header()
//
//		for _, st := range stores {
//			st.PrintInfo(w)
//		}
//		w.Flush()
//		return nil
//	}
