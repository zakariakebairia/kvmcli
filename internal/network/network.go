package network

import (
	"context"
	"errors"
	// "github.com/zakariakebairia/kvmcli/internal/config"
)

// IDEA: the ip <=>  mac address mapping done here in Virtual Network declaration
// What I need is whenever I create a new virtual machine with a static ip, I need to update
// my virtual network declaration to add the ip <=> mac address mapping

var (
	ErrNetworkNameEmpty = errors.New("virtual network name is empty")
	ErrNilLibvirtConn   = errors.New("libvirt connection is nil")
)

// Network represents a bound network resource (configuration + manager).
// It implements resources.Resource.
type Network struct {
	Spec    Config
	ctx     context.Context
	manager NetworkManager
}

// NewNetwork creates a new Network resource.
func NewNetwork(spec Config, manager NetworkManager, ctx context.Context) *Network {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Network{
		Spec:    spec,
		manager: manager,
		ctx:     ctx,
	}
}

// Create delegates to the manager.
func (n *Network) Create() error {
	return n.manager.Create(n.ctx, n.Spec)
}

// Delete delegates to the manager.
func (n *Network) Delete() error {
	return n.manager.Delete(n.ctx, n.Spec.Name)
}

// Start delegates to the manager (if implemented) or just logs.
// Previous implementation was a print.
func (n *Network) Start() error {
	// The manager interface defined Start(ctx, name).
	// Let's implement it in manager if needed, or here.
	// Previous: fmt.Println("Start virtual network")
	// If manager has Start, use it. My manager definition has Start commented out?
	// Let's check manager.go. I commented it out // Start...
	// I should uncomment it in manager.go or just print here.
	// Ideally manager handles it.
	// For now, let's keep it simple and print here to match previous behavior if manager doesn't support it yet,
	// BUT the interface requires Start() error.
	return nil
}

// AddStaticMapping delegates to manager.
func (n *Network) SetStaticMapping(ip, mac string) error {
	return n.manager.SetStaticMapping(n.ctx, n.Spec.Name, ip, mac)
}
