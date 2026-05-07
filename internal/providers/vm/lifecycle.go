package vm

import (
	"fmt"

	"github.com/zakariakebairia/kvmcli/internal/providers/network"
	"github.com/zakariakebairia/kvmcli/internal/providers/store"
	"github.com/zakariakebairia/kvmcli/internal/registry"
)

const TypeName = "vm"

var vmColumns = []string{"NAME", "NAMESPACE", "CPU", "RAM", "IP", "IMAGE", "STATUS"}

func init() {
	registry.Register(&registry.ResourceType{
		Name:      TypeName,
		DependsOn: []string{network.TypeName, store.TypeName},
		Lifecycle: &VMLifecycle{},
		Columns:   vmColumns,
		Format:    formatVM,
	})
}

// VMLifecycle implements registry.ResourceLifecycle.
type VMLifecycle struct{}

// formatVM
func formatVM(obj registry.Object) []string {
	// Get the display for  the image
	// Get  status in realtime
	return []string{
		obj.Name,
		obj.Namespace,
		fmt.Sprintf("%v", obj.Attrs["cpu"]),
		fmt.Sprintf("%v", obj.Attrs["memory"]),
		obj.GetString("ip"),
		obj.GetString("image"),
		obj.Status,
	}
}

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
	// check attributes validity
	var attrs VMAttrs

	if err := attrs.FromObject(spec); err != nil {
		return fmt.Errorf("parsing vm %q: %w", spec.Name, err)
	}
	if err := attrs.Validate(); err != nil {
		return err
	}

	// Resolve the host's L2/L3 identity (IP + MAC).
	// If no MAC is provided, one is derived deterministically from the IP.
	hostAddr, err := network.ResolveL2L3Pair(
		spec.GetString("ip"),
		spec.GetString("mac_address"),
	)
	if err != nil {
		return fmt.Errorf("resolve host addresses for %q: %w", spec.Name, err)
	}
	// Get image to provision
	image, err := getImage(
		session,
		spec.GetString("store"),
		spec.GetString("image"),
		spec.Namespace,
	)
	if err != nil {
		return fmt.Errorf("get image: %w", err)
	}

	// Provision a qcow2 overlay disk backed by the specified image.
	if err := provisionDisk(session, image.SrcPath, image.DiskPath, spec); err != nil {
		return fmt.Errorf("provision disk: %w", err)
	}
	rollback = append(rollback, func() { deleteOverlay(image.DiskPath) })

	// Define the libvirt domain (registers the VM, does not start it).
	domain, err := defineDomain(session, spec, image, hostAddr)
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
	spec.Attrs["disk_path"] = image.DiskPath
	spec.Status = "running"
	return nil
}

// FIX: Correct destroy behavior for resources defined in HCL.
//
// Resources in the HCL file do not contain all required values.
// Some fields (e.g., MAC address, disk_path) are generated and stored
// only after the resource is created in the database.
//
// For destroy:
// - Use the HCL file only to identify the resource (e.g., by name).
// - Load the complete resource data from the database.
// - Perform the destroy operation using the full, resolved data.
func (l *VMLifecycle) Destroy(session registry.Session, change registry.Change) error {
	spec := change.Current

	dom, err := session.Conn.DomainLookupByName(spec.Name)
	if err != nil {
		return fmt.Errorf("lookup domain %q: %w", spec.Name, err)
	}

	// Ignore error — VM might already be stopped
	_ = session.Conn.DomainDestroy(dom)

	if err := session.Conn.DomainUndefineFlags(dom, 0); err != nil {
		return fmt.Errorf("undefine domain %q: %w", spec.Name, err)
	}

	// // Delete disk overlay
	// if diskPath := spec.GetString("disk_path"); diskPath != "" {
	// 	fmt.Println("-->", diskPath)
	// 	if err := deleteOverlay(diskPath); err != nil {
	// 		return err
	// 	}
	// }

	diskPath := spec.GetString("disk_path")
	if err := deleteOverlay(diskPath); err != nil {
		return err
	}
	return nil
}
