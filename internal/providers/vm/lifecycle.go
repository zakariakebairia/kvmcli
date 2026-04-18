package vm

import (
	"fmt"

	logger "github.com/zakariakebairia/kvmcli/internal/logger"
	"github.com/zakariakebairia/kvmcli/internal/providers/network"
	"github.com/zakariakebairia/kvmcli/internal/registry"
)

// type MACAddress = net.HardwareAddr

func init() {
	registry.Register(&registry.ResourceType{
		Name:      "vm",
		DependsOn: []string{"network", "store"},
		Lifecycle: &VMLifecycle{},
		Columns:   []string{"NAME", "NAMESPACE", "CPU", "RAM", "IP", "IMAGE", "STATUS"},
		Format: func(object registry.Object) []string {
			return []string{
				object.Name,
				object.Namespace,
				fmt.Sprintf("%v", object.Attrs["cpu"]),
				fmt.Sprintf("%v", object.Attrs["memory"]),
				object.GetString("ip"),
				object.GetString("image"),
				object.Status,
			}
		},
	})
}

// VMLifecycle implements registry.ResourceLifecycle.
type VMLifecycle struct{}

func (l *VMLifecycle) Plan(desired, current *registry.Object) (registry.Action, error) {
	if current == nil && desired != nil {
		return registry.ActionCreate, nil
	}
	if current != nil && desired == nil {
		return registry.ActionDelete, nil
	}

	if current != nil && desired != nil {
		return registry.ActionUpdate, nil
	}
	return registry.ActionNone, nil
}

func (vm *VMLifecycle) Apply(session registry.Session, change registry.Change) (err error) {
	// rollback is a LIFO stack of cleanup functions.
	// It runs automatically on any error via the deferred func below.
	var rollback []func()
	defer func() {
		if err != nil {
			for i := len(rollback) - 1; i >= 0; i-- {
				rollback[i]()
			}
		}
	}()

	spec := change.Desired

	// Resolve the host's L2/L3 identity (IP + MAC).
	// If no MAC is provided, one is derived deterministically from the IP.
	hostAddr, err := network.ResolveL2L3Pair(
		spec.GetString("ip"),
		spec.GetString("mac"),
	)
	if err != nil {
		return fmt.Errorf("resolve host addresses for %q: %w", spec.Name, err)
	}

	// Provision a qcow2 overlay disk backed by the specified image.
	diskPath, err := provisionDisk(session, spec)
	if err != nil {
		return fmt.Errorf("provision disk: %w", err)
	}
	rollback = append(rollback, func() { deleteOverlay(diskPath) })

	// Define the libvirt domain (registers the VM, does not start it).
	domain, err := defineDomain(session, spec, diskPath, hostAddr)
	if err != nil {
		return fmt.Errorf("define domain %q: %w", spec.Name, err)
	}
	rollback = append(rollback, func() { session.Conn.DomainUndefineFlags(domain, 0) })

	// Register a static DHCP mapping so the VM always gets the same IP.
	if err = network.SetStaticMapping(session, spec, hostAddr); err != nil {
		return fmt.Errorf("set static DHCP mapping for %q: %w", spec.Name, err)
	}
	// rollback = append(rollback, func() { network.RemoveStaticMapping(session, hostAddr) })

	// Start the domain (boots the VM).
	if err = createDomain(session, domain); err != nil {
		return fmt.Errorf("start domain %q: %w", spec.Name, err)
	}

	// Persist computed values back into the spec so the engine can save them.
	spec.Attrs["mac_address"] = hostAddr.MAC.String()
	spec.Attrs["disk_path"] = diskPath
	spec.Status = "running"
	return nil
}

func (l *VMLifecycle) Destroy(session registry.Session, current registry.Object) error {
	dom, err := session.Conn.DomainLookupByName(current.Name)
	if err != nil {
		return fmt.Errorf("lookup domain %q: %w", current.Name, err)
	}

	// Ignore error — VM might already be stopped
	_ = session.Conn.DomainDestroy(dom)

	if err := session.Conn.DomainUndefineFlags(dom, 0); err != nil {
		return fmt.Errorf("undefine domain %q: %w", current.Name, err)
	}

	// Delete disk overlay
	if diskPath := current.GetString("disk_path"); diskPath != "" {
		if err := deleteOverlay(diskPath); err != nil {
			logger.Warnf("failed to delete disk %s: %v", diskPath, err)
		}
	}

	return nil
}
