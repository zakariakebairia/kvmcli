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
	// identity, err := vm.resolveIdentity(spec)
	// disk, err     := vm.provisionDisk(session, spec)
	// domain, err   := vm.defineDomain(session, spec, disk, identity)
	// err            = vm.registerNetwork(session, spec, identity)
	// err            = vm.startDomain(session, domain)
	// vm.persistState(spec, identity, disk)

	var rollback []func()
	defer func() {
		if err != nil {
			for i := len(rollback) - 1; i >= 0; i-- {
				rollback[i]()
			}
		}
	}()

	spec := change.Desired
	hostAddr, err := network.ResolveL2L3Pair(spec.GetString("ip"), spec.GetString("mac"))
	if err != nil {
		return fmt.Errorf("resolve host addresses for %q: %w", spec.Name, err)
	}

	dest, err := provisionDisk(session, spec)
	if err != nil {
		return fmt.Errorf("provision disk: %w", err)
	}
	rollback = append(rollback, func() { deleteOverlay(session.Ctx, dest) })

	// defineDomain
	domain, err := defineDomain(session, spec, dest, hostAddr)
	if err != nil {
		return fmt.Errorf("define domain: %w", err)
	}
	rollback = append(rollback, func() { session.Conn.DomainUndefineFlags(domain, 0) })
	// TODO: defineNetwork
	// FIX: needs a rollback function
	if err := network.SetStaticMapping(
		session,
		spec,
		hostAddr,
	); err != nil {
		return fmt.Errorf("set static mapping: %w", err)
	}
	// rollback = append(rollback, func() { network.RemoveStaticMapping(session, hostBinding) })

	// TODO: createDomain
	if err := createDomain(session, domain); err != nil {
		return fmt.Errorf("create domain: %w", err)
	}
	// Persist computed values into attrs so the engine saves them
	spec.Attrs["mac_address"] = hostAddr.MAC.String()
	spec.Attrs["disk_path"] = dest
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
		if err := deleteOverlay(session.Ctx, diskPath); err != nil {
			logger.Warnf("failed to delete disk %s: %v", diskPath, err)
		}
	}

	return nil
}

// func (l *VMLifecycle) Apply(session registry.Session, change registry.Change) error {
// 	spec := change.Desired
//
// 	// Resolve MAC address
// 	// macAddress, err := network.ResolveMAC("02:aa:bb", attrStr(*spec, "ip"), attrStr(*spec, "mac"))
// 	macAddress, err := network.IP2MAC(spec.GetString("ip"))
// 	if err != nil {
// 		return fmt.Errorf("resolve mac for %q: %w", spec.Name, err)
// 	}
//
// 	// TODO:  Look up image info from the store's state
// 	// I will use the new Pool/Volume object
// 	// artifactsPath, imagesPath, imageFile, osProfile, err := lookupImage(
// 	// 	session,
// 	// 	spec.GetString("image"),
// 	// )
// 	image, err := getImage(session, spec.GetString("image"))
// 	if err != nil {
// 		return fmt.Errorf("lookup image: %w", err)
// 	}
//
// 	src := filepath.Join(image.ArtifactsPath, image.ImageFile)
// 	dest := filepath.Join(image.ImagesPath, spec.Name+".qcow2")
//
// 	// Step 3: Create disk overlay
// 	if err := createOverlay(session.Ctx, src, dest); err != nil {
// 		return fmt.Errorf("create disk overlay: %w", err)
// 	}
//
// 	// Step 4: Build XML and define domain
// 	netName := spec.GetString("network")
// 	xmlConfig, err := buildDomainXML(spec, dest, netName, macAddress, image.OsProfile)
// 	if err != nil {
// 		deleteOverlay(session.Ctx, dest)
// 		return fmt.Errorf("build XML: %w", err)
// 	}
//
// 	dom, err := session.Conn.DomainDefineXML(xmlConfig)
// 	if err != nil {
// 		deleteOverlay(session.Ctx, dest)
// 		return fmt.Errorf("define domain %q: %w", spec.Name, err)
// 	}
//
// 	// Step 5: Set static IP mapping on the network
// 	// TODO: I need to work on this parsing functions below
// 	// I think that the parsing needs to be on config loading stage
// 	if ip := spec.GetString("ip"); ip != "" {
// 		parsedIP := net.ParseIP(ip)
// 		parsedMAC, err := net.ParseMAC(macAddress)
// 		if err != nil {
// 			return fmt.Errorf("invalid MAC %q: %w", macAddress, err)
// 		}
// 		networkManager := network.NewNetworkManager(session.Conn, session.DB)
// 		if err := networkManager.SetStaticMapping(
// 			netName,
// 			parsedIP,
// 			parsedMAC,
// 		); err != nil {
// 			session.Conn.DomainUndefineFlags(dom, 0)
// 			deleteOverlay(session.Ctx, dest)
// 			return fmt.Errorf("set static mapping: %w", err)
// 		}
// 	}
//
// 	// Step 6: Start the VM
// 	if err := session.Conn.DomainCreate(dom); err != nil {
// 		session.Conn.DomainUndefineFlags(dom, 0)
// 		deleteOverlay(session.Ctx, dest)
// 		return fmt.Errorf("start domain %q: %w", spec.Name, err)
// 	}
//
// 	// Persist computed values into attrs so the engine saves them
// 	spec.Attrs["mac_address"] = macAddress
// 	spec.Attrs["disk_path"] = dest
// 	spec.Status = "running"
//
// 	return nil
// }
