package vms

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/zakariakebairia/kvmcli/internal/database"
	db "github.com/zakariakebairia/kvmcli/internal/database"
	log "github.com/zakariakebairia/kvmcli/internal/logger"
)

const (
	deviceName = "vda"
	// qemu
	defaultQemuTimeout = 30 * time.Second
	qemuImgCmd         = "qemu-img"
)

// NewVMRecord constructs a new VM record from the provided VM configuration.
// then creates a database record for the virtual machine.
func NewVirtualMachineRecord(
	vm *VirtualMachine,
) (*db.VirtualMachine, error) {
	store, err := vm.fetchStore()
	if err != nil {
		return nil, fmt.Errorf("failed to get store: %w", err)
	}

	// Verify image exists in store
	if _, err := database.GetImage(vm.ctx, vm.db, vm.Spec.Image); err != nil {
		return nil, fmt.Errorf(
			"image %q not found in store %q: %w",
			vm.Spec.Image,
			vm.Spec.Store,
			err,
		)
	}

	networkID, err := db.GetNetworkIDByName(vm.ctx, vm.db, vm.Spec.NetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get network ID: %w", err)
	}

	storeID, err := db.GetStoreIDByName(vm.ctx, vm.db, vm.Spec.Store)
	if err != nil {
		return nil, fmt.Errorf("failed to get network ID: %w", err)
	}
	diskPath := filepath.Join(store.ImagesPath, vm.Spec.Name+".qcow2")

	return &db.VirtualMachine{
		Name:      vm.Spec.Name,
		Namespace: vm.Spec.Namespace,
		Labels:    vm.Spec.Labels,
		CPU:       vm.Spec.CPU,
		RAM:       vm.Spec.Memory,
		// DiskSize:   vm.Spec.Disk.Size,
		DiskSize:   vm.Spec.Disk,
		DiskPath:   diskPath,
		Image:      vm.Spec.Image,
		MacAddress: vm.Spec.MAC,
		IP:         vm.Spec.IP,
		NetworkID:  networkID,
		StoreID:    storeID,
		CreatedAt:  time.Now(),
	}, nil
}

func (vm *VirtualMachine) rollback(cleanups []func() error, step string, originError error) error {
	for _, fn := range cleanups {
		if err := fn(); err != nil {
			log.Warnf("rollback failed, step %s, err %s", step, err)
		}
	}

	return fmt.Errorf("failed at %s: %w", step, originError)
}
