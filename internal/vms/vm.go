package vms

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/digitalocean/go-libvirt"
	db "github.com/zakariakebairia/kvmcli/internal/database"
)

// Error definitions
var (
	ErrNilLibvirtConn   = errors.New("libvirt connection is nil")
	ErrNilDBConn        = errors.New("database connection is nil")
	ErrNilDiskManager   = errors.New("disk manager is nil")
	ErrNilDomainManager = errors.New("domain manager is nil")
)

// VirtualMachine represents a Config with injected dependencies.
// I need to remove the config, I need to make it as independent as possible
// and use the config as extra
// We need config only in creation, and nothing beyound that
type VirtualMachine struct {
	ctx    context.Context
	conn   *libvirt.Libvirt
	db     *sql.DB
	Spec   Config
	disk   DiskManager
	domain DomainManager
	// network NetworkManager
}

// getStore retrieves the store record for the Config.
// getStore retrieves the store record for the Config.
func (vm *VirtualMachine) fetchStore() (*db.Store, error) {
	var store db.Store
	// Fetch the full store record using the name from configuration
	if err := store.GetRecord(vm.ctx, vm.db, vm.Spec.Store); err != nil {
		return nil, fmt.Errorf("failed to get store record for %q: %w", vm.Spec.Store, err)
	}

	return &store, nil
}

// VirtualMachineOption is a functional option for configuring a VirtualMachine.
type VirtualMachineOption func(*VirtualMachine)

// WithLibvirtConnection injects a non-nil Libvirt connection.
func WithLibvirtConnection(conn *libvirt.Libvirt) VirtualMachineOption {
	return func(vm *VirtualMachine) {
		vm.conn = conn
	}
}

// WithDatabaseConnection injects a non-nil *sql.DB.
func WithDatabaseConnection(db *sql.DB) VirtualMachineOption {
	return func(vm *VirtualMachine) {
		vm.db = db
	}
}

// WithContext injects a context.Context; if nil is passed, a background context is set.
func WithContext(ctx context.Context) VirtualMachineOption {
	return func(vm *VirtualMachine) {
		if ctx == nil {
			vm.ctx = context.Background()
		} else {
			vm.ctx = ctx
		}
	}
}

// New option to inject a DiskManager:
func WithDiskManager(d DiskManager) VirtualMachineOption {
	return func(vm *VirtualMachine) {
		vm.disk = d
	}
}

// New option to inject a DomainManager:
func WithDomainManager(d DomainManager) VirtualMachineOption {
	return func(vm *VirtualMachine) {
		vm.domain = d
	}
}

// NewVirtualMachine constructs a Config, applies options, and validates dependencies.
func NewVirtualMachine(
	cfg Config,
	opts ...VirtualMachineOption,
) (*VirtualMachine, error) {
	vm := &VirtualMachine{
		Spec: cfg,
		ctx:  context.Background(),
	}

	// apply options
	for _, opt := range opts {
		opt(vm)
	}

	// wire defaults if not provided
	// if vm.disk == nil {
	// 	// user must provide via option; no hard-coded defaults here
	// 	return nil, ErrNilDiskManager
	// }
	// if vm.domain == nil {
	// 	return nil, ErrNilDomainManager
	// }

	// validate core dependencies
	if vm.conn == nil {
		return nil, ErrNilLibvirtConn
	}
	if vm.db == nil {
		return nil, ErrNilDBConn
	}

	if vm.disk == nil {
		store, err := vm.fetchStore()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch store for disk configuration: %w", err)
		}

		vm.disk = &QemuDiskManager{
			QemuImgPath:    "qemu-img", // Default to system path
			BaseImagesPath: store.ArtifactsPath,
			DestImagesPath: store.ImagesPath,
			Timeout:        10 * time.Second,
		}
	}
	vm.domain = NewLibvirtDomainManager(vm.conn)

	return vm, nil
}
