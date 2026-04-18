package vm

import (
	"encoding/xml"
	"fmt"

	"github.com/digitalocean/go-libvirt"
	"github.com/zakariakebairia/kvmcli/internal/providers/network"
	"github.com/zakariakebairia/kvmcli/internal/registry"
	"github.com/zakariakebairia/kvmcli/internal/templates"
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
	diskPath string,
	hostAddr *network.HostAddr,
) (domain libvirt.Domain, err error) {
	// Get network name
	networkName := spec.GetString("network")
	// Build xml
	xml, err := buildDomainXML(
		spec,
		diskPath,
		networkName,
		hostAddr.MAC.String(),
		"https://rockylinux.org/rocky/9",
	)
	if err != nil {
		return domain, fmt.Errorf("build XML: %w", err)
	}
	domain, err = session.Conn.DomainDefineXML(xml)
	if err != nil {
		// TODO: fix this
		return domain, fmt.Errorf("define domain %q: %w", spec.Name, err)
	}
	return domain, nil
}

func createDomain(session registry.Session, domain libvirt.Domain) error {
	if err := session.Conn.DomainCreate(domain); err != nil {
		return fmt.Errorf("start domain %q: %w", domain.Name, err)
	}
	return nil
}
