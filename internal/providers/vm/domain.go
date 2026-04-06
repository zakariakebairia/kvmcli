package vm

import (
	"fmt"

	"github.com/digitalocean/go-libvirt"
)

// Start powers on a VM domain by name.
func Start(conn *libvirt.Libvirt, name string) error {
	dom, err := conn.DomainLookupByName(name)
	if err != nil {
		return fmt.Errorf("lookup domain %q: %w", name, err)
	}
	if err := conn.DomainCreate(dom); err != nil {
		return fmt.Errorf("start domain %q: %w", name, err)
	}
	return nil
}

// Stop gracefully shuts down a VM domain by name.
func Stop(conn *libvirt.Libvirt, name string) error {
	dom, err := conn.DomainLookupByName(name)
	if err != nil {
		return fmt.Errorf("lookup domain %q: %w", name, err)
	}
	if err := conn.DomainShutdown(dom); err != nil {
		return fmt.Errorf("shutdown domain %q: %w", name, err)
	}
	return nil
}
