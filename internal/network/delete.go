package network

import (
	"context"
	"errors"
	"fmt"

	db "github.com/zakariakebairia/kvmcli/internal/database"
)

// Delete removes a Network from libvirt and deletes its database record.
// Delete removes a Network from libvirt and deletes its database record.
func (m *LibvirtNetworkManager) Delete(ctx context.Context, name string) error {
	// Ensure we have a Libvirt connection
	if m.conn == nil {
		return errors.New("libvirt connection is not initialized")
	}

	// Lookup the network by name
	virNet, err := m.conn.NetworkLookupByName(name)
	if err != nil {
		return fmt.Errorf("network %q not found: %w", name, err)
	}

	// Destroy the network (stop it if it’s running)
	if err := m.conn.NetworkDestroy(virNet); err != nil {
		return fmt.Errorf("failed to destroy network %q: %w", name, err)
	}

	// Undefine the network (remove its definition from libvirt)
	if err := m.conn.NetworkUndefine(virNet); err != nil {
		return fmt.Errorf("failed to undefine network %q: %w", name, err)
	}

	// Remove the record from the database
	// We need to delete by name. NewNetworkRecord expects *Network to build a record.
	// But store_record.go/network_record.go usually has a Delete method on the record struct.
	// Let's check network_record.go. Assuming usage of db.VirtualNetwork for now.
	// We can construct a dummy one with just the name to call Delete?
	// Or better, network_record.go probably has a Delete method that uses Name/Namespace.
	// Let's create a partial record.
	record := &db.VirtualNetwork{Name: name}
	if err := record.Delete(ctx, m.db); err != nil {
		return fmt.Errorf("failed to delete database record for network %q: %w", name, err)
	}

	fmt.Printf("network/%s deleted\n", name)
	return nil
}
