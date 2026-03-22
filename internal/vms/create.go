package vms

import (
	"fmt"
	"path/filepath"

	"github.com/zakariakebairia/kvmcli/internal/database"
	"github.com/zakariakebairia/kvmcli/internal/network"
)

// Create Virtual Machine
func (vm *VirtualMachine) Create() error {
	// Resolve MAC address (explicit in config, otherwise derived from IP).
	macAddress, err := network.ResolveMAC("02:aa:bb", vm.Spec.IP, vm.Spec.MAC)
	if err != nil {
		return fmt.Errorf("resolve mac for %q: %w", vm.Spec.Name, err)
	}
	// Initiliaze a new vm record
	record, err := NewVirtualMachineRecord(vm)
	if err != nil {
		return fmt.Errorf("can't Initiliaze a new vm: %w", err)
	}

	// store, img, err := vm.fetchStoreAndImage(vm.Config.Spec.Image)
	// if err != nil {
	// 	return fmt.Errorf("fetch store and image: %w", err)
	// }
	img, err := database.GetImage(vm.ctx, vm.db, vm.Spec.Image)
	if err != nil {
		return fmt.Errorf("fetch store and image: %w", err)
	}

	src := filepath.Join(img.ArtifactsPath, img.ImageFile)
	dest := filepath.Join(img.ImagesPath, vm.Spec.Name+".qcow2")

	// this is a slice that collection all the cleanup functions
	// so that when an error happens in each step, a proper rollback
	// will be executed
	var cleanups []func() error

	// artifactsPath, imagesPath := vm.disk.Paths()
	// src := fmt.Sprintf("%s/%s", artifactsPath, vm.Config.Spec.Image)
	// dest := fmt.Sprintf("%s/%s.qcow2", imagesPath, vm.Config.Metadata.Name)
	if err := vm.disk.CreateOverlay(vm.ctx, src, dest); err != nil {
		return fmt.Errorf("create disk overlay: %w", err)
	}
	//
	cleanups = append(cleanups, func() error {
		return vm.disk.DeleteOverlay(vm.ctx, dest)
	})

	// Generate the libvirt XML configuration
	xmlConfig, err := vm.domain.BuildXML(vm.ctx, vm.db, vm.Spec)
	if err != nil {
		return vm.rollback(cleanups, "build XML", err)
	}

	// Step 3: Define and start the VM
	if err := vm.domain.Define(vm.ctx, xmlConfig); err != nil {
		return vm.rollback(cleanups, "define domain", err)
	}

	cleanups = append(cleanups, func() error {
		return vm.domain.Undefine(vm.ctx, vm.Spec.Name)
	})

	// Step 4: Add static IP mapping if configured
	if vm.Spec.IP != "" {
		nm := network.NewLibvirtNetworkManager(vm.conn, vm.db)
		// We're using NetName which connects to the network name in config
		// if err := nm.SetStaticMapping(vm.ctx, vm.Spec.NetName, vm.Spec.IP, vm.Spec.MAC); err != nil {
		if err := nm.SetStaticMapping(vm.ctx, vm.Spec.NetName, vm.Spec.IP, macAddress); err != nil {
			// We might want to warn instead of fail, or fail.
			// If we fail, we should rollback (undefine domain).
			return vm.rollback(cleanups, "add static ip mapping", err)
		}
	}

	if err := vm.domain.Start(vm.ctx, vm.Spec.Name); err != nil {
		return vm.rollback(cleanups, "start domain", err)
	}

	// cleanups = append(cleanups, vm.domain.)
	cleanups = append(cleanups, func() error {
		return vm.domain.Stop(vm.ctx, vm.Spec.Name)
	})

	// Insert the vm record
	if err = record.Insert(vm.ctx, vm.db); err != nil {
		return vm.rollback(cleanups, "insert record", err)
	}

	fmt.Printf("vm/%s created\n", vm.Spec.Name)
	return nil
}
