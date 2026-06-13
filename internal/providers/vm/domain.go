package vm

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/zakariakebairia/kvmcli/internal/database"
	"github.com/zakariakebairia/kvmcli/internal/providers/network"
	"github.com/zakariakebairia/kvmcli/internal/registry"
	"github.com/zakariakebairia/kvmcli/internal/templates"
)

// Start powers on a VM domain by name.
func Start(session registry.Session, name string) error {
	domain, err := session.Conn.DomainLookupByName(name)
	if err != nil {
		return fmt.Errorf("lookup domain %q: %w", name, err)
	}
	if err := session.Conn.DomainCreate(domain); err != nil {
		// return err
		return fmt.Errorf("start domain %q: %w", name, err)
	}
	dbHandler := database.NewDBHandler(session.DB)
	object, err := dbHandler.Get(session.Ctx, TypeName, name, "infra")

	status, err := GetVMStatus(session, name)
	if err != nil {
		return fmt.Errorf("get vm status: %w", err)
	}
	object.Status = status
	if err := dbHandler.Put(session.Ctx, object); err != nil {
		return err
	}
	return nil
}

func GetVMAge(object registry.Object) (string, error) {
	if object.CreatedAt == "" {
		return "", fmt.Errorf("created_at is empty for vm %s", object.Name)
	}

	t, err := time.Parse(time.RFC3339, object.CreatedAt)
	if err != nil {
		return "", fmt.Errorf("parsing created_at: %w", err)
	}

	return formatAge(time.Since(t)), nil
}

func formatAge(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func GetVMStatus(session registry.Session, name string) (string, error) {
	domain, err := session.Conn.DomainLookupByName(name)
	if err != nil {
		return "", fmt.Errorf("lookup domain %q: %w", name, err)
	}
	state, _, err := session.Conn.DomainGetState(domain, 0)
	if err != nil {
		return "", fmt.Errorf("get state: %w", err)
	}
	switch libvirt.DomainState(state) {
	case libvirt.DomainRunning:
		return "Running", nil

	case libvirt.DomainBlocked:
		return "Blocked", nil

	case libvirt.DomainPaused:
		return "Paused", nil

	case libvirt.DomainShutdown:
		return "Shutdown", nil

	case libvirt.DomainShutoff:
		return "Shutoff", nil

	case libvirt.DomainCrashed:
		return "Crashed", nil

	case libvirt.DomainPmsuspended:
		return "Suspended", nil

	default:
		return "Unknown", nil
	}
}

func Stop(session registry.Session, name string) error {
	dbHandler := database.NewDBHandler(session.DB)
	object, err := dbHandler.Get(session.Ctx, TypeName, name, "infra")
	if err != nil {
		return fmt.Errorf("get operation: %w", err)
	}
	if object.Status == "stopped" {
		return fmt.Errorf("vm %q is already stopped", name)
	}
	domain, err := session.Conn.DomainLookupByName(name)
	if err != nil {
		return fmt.Errorf("lookup domain %q: %w", name, err)
	}
	if err := session.Conn.DomainShutdown(domain); err != nil {
		return fmt.Errorf("shutdown domain %q: %w", name, err)
	}

	status, err := GetVMStatus(session, name)
	if err != nil {
		return fmt.Errorf("get vm status: %w", err)
	}

	object.Status = status

	if err := dbHandler.Put(session.Ctx, object); err != nil {
		return err
	}
	return nil
}

// buildDomainXML generates the libvirt XML for a VM domain.
func buildDomainXML(
	spec *registry.Object,
	diskPath, netName, macAddress, osProfile string,
) (string, error) {
	cpu := spec.GetInt("cpu")
	memory := spec.GetInt("memory")

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

func defineDomain(
	session registry.Session,
	spec *registry.Object,
	image *Image,
	hostAddr *network.HostAddr,
) (domain libvirt.Domain, err error) {
	// Get network name
	networkName := spec.GetString("network")

	// Build xml
	xml, err := buildDomainXML(
		spec,
		image.DiskPath,
		networkName,
		hostAddr.MAC.String(),
		image.OsProfile,
	)
	if err != nil {
		return domain, fmt.Errorf("build XML: %w", err)
	}
	domain, err = session.Conn.DomainDefineXML(xml)
	// Naked return: implicitly returns the named result parameters declared in the function signature.
	if err != nil {
		return
	}
	return domain, nil
}

func createDomain(session registry.Session, domain libvirt.Domain) error {
	if err := session.Conn.DomainCreate(domain); err != nil {
		return fmt.Errorf("start domain %q: %w", domain.Name, err)
	}
	return nil
}
