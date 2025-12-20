package operations

import (
	"context"
	"fmt"
	"time"

	"github.com/kebairia/kvmcli/internal/network"
	"github.com/kebairia/kvmcli/internal/store"
	"github.com/kebairia/kvmcli/internal/vms"
)

// -------------------------------------------
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
func ListAll() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	operator, err := NewOperator(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to initialize operator: %w", err)
	}
	defer operator.Close()

	virtualMachines, err := vms.GetVirtualMachines(operator.ctx, operator.db, operator.conn)
	if err != nil {
		return fmt.Errorf("failed to retrieve VMs: %w", err)
	}

	info := &vms.VirtualMachineInfo{}
	w := info.Header()

	for _, virtualMachine := range virtualMachines {
		virtualMachine.PrintInfo(w)
	}
	w.Flush()
	return nil
}

func ListAllNetworks() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	operator, err := NewOperator(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to initialize operator: %w", err)
	}
	defer operator.Close()

	virtualNetworks, err := network.GetVirtualNetworks(operator.ctx, operator.db, operator.conn)
	if err != nil {
		return fmt.Errorf("failed to retrieve networks: %w", err)
	}

	info := &network.VirtualNetworkInfo{}
	w := info.Header()

	for _, virtualNetwork := range virtualNetworks {
		virtualNetwork.PrintInfo(w)
	}
	w.Flush()
	return nil
}

// ListAllStores lists all stores in the database.
func ListAllStores() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	operator, err := NewOperator(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to initialize operator: %w", err)
	}
	defer operator.Close()

	stores, err := store.GetStores(operator.ctx, operator.db)
	if err != nil {
		return fmt.Errorf("failed to retrieve stores: %w", err)
	}

	info := &store.StoreInfo{}
	w := info.Header()

	for _, st := range stores {
		st.PrintInfo(w)
	}
	w.Flush()
	return nil
}
