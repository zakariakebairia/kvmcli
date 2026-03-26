package vm

import (
	"encoding/xml"
	"fmt"
	"path/filepath"

	"github.com/zakariakebairia/kvmcli/internal/network"
	"github.com/zakariakebairia/kvmcli/internal/registry"
	"github.com/zakariakebairia/kvmcli/internal/templates"
)

func init() {
	registry.Register(&registry.ResourceType{
		Name:      "vm",
		DependsOn: []string{"network", "store"},
		Lifecycle: &VMLifecycle{},
		Columns:   []string{"NAME", "NAMESPACE", "CPU", "RAM", "IP", "IMAGE", "STATUS"},
		Format: func(s registry.Object) []string {
			return []string{
				s.Name,
				s.Namespace,
				fmt.Sprintf("%v", s.Attrs["cpu"]),
				fmt.Sprintf("%v", s.Attrs["memory"]),
				attrStr(s, "ip"),
				attrStr(s, "image"),
				s.Status,
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
	return registry.ActionNone, nil
}

func (l *VMLifecycle) Apply(session registry.Session, change registry.Change) error {
	spec := change.Desired

	// Step 1: Resolve MAC address
	macAddress, err := network.ResolveMAC("02:aa:bb", attrStr(*spec, "ip"), attrStr(*spec, "mac"))
	if err != nil {
		return fmt.Errorf("resolve mac for %q: %w", spec.Name, err)
	}

	// Step 2: Look up image info from the store's state
	artifactsPath, imagesPath, imageFile, osProfile, err := lookupImage(session, attrStr(*spec, "image"))
	if err != nil {
		return fmt.Errorf("lookup image: %w", err)
	}

	src := filepath.Join(artifactsPath, imageFile)
	dest := filepath.Join(imagesPath, spec.Name+".qcow2")

	// Step 3: Create disk overlay
	disk := newDiskManager()
	if err := disk.CreateOverlay(session.Ctx, src, dest); err != nil {
		return fmt.Errorf("create disk overlay: %w", err)
	}

	// Step 4: Build XML and define domain
	netName := attrStr(*spec, "network")
	xmlConfig, err := buildDomainXML(spec, dest, netName, macAddress, osProfile)
	if err != nil {
		disk.DeleteOverlay(session.Ctx, dest)
		return fmt.Errorf("build XML: %w", err)
	}

	dom, err := session.Conn.DomainDefineXML(xmlConfig)
	if err != nil {
		disk.DeleteOverlay(session.Ctx, dest)
		return fmt.Errorf("define domain %q: %w", spec.Name, err)
	}

	// Step 5: Set static IP mapping on the network
	if ip := attrStr(*spec, "ip"); ip != "" {
		networkManager := network.NewLibvirtNetworkManager(session.Conn, session.DB)
		if err := networkManager.SetStaticMapping(session.Ctx, netName, ip, macAddress); err != nil {
			session.Conn.DomainUndefineFlags(dom, 0)
			disk.DeleteOverlay(session.Ctx, dest)
			return fmt.Errorf("set static mapping: %w", err)
		}
	}

	// Step 6: Start the VM
	if err := session.Conn.DomainCreate(dom); err != nil {
		session.Conn.DomainUndefineFlags(dom, 0)
		disk.DeleteOverlay(session.Ctx, dest)
		return fmt.Errorf("start domain %q: %w", spec.Name, err)
	}

	// Persist computed values into attrs so the engine saves them
	spec.Attrs["mac_address"] = macAddress
	spec.Attrs["disk_path"] = dest
	spec.Status = "running"

	fmt.Printf("vm/%s created\n", spec.Name)
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
	if diskPath := attrStr(current, "disk_path"); diskPath != "" {
		disk := newDiskManager()
		if err := disk.DeleteOverlay(session.Ctx, diskPath); err != nil {
			fmt.Printf("warning: failed to delete disk %s: %v\n", diskPath, err)
		}
	}

	fmt.Printf("vm/%s deleted\n", current.Name)
	return nil
}

// buildDomainXML generates the libvirt XML for a VM domain.
func buildDomainXML(spec *registry.Object, diskPath, netName, macAddress, osProfile string) (string, error) {
	cpu := attrInt(*spec, "cpu")
	memory := attrInt(*spec, "memory")

	domain := templates.NewDomain(
		spec.Name,
		memory,
		cpu,
		diskPath,
		netName,
		macAddress,
		osProfile,
	)

	xmlConfig, err := domain.GenerateXML()
	if err != nil {
		return "", fmt.Errorf("generate XML for vm %s: %w", spec.Name, err)
	}

	return xml.Header + string(xmlConfig), nil
}
