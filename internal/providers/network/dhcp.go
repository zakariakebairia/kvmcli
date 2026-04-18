package network

import (
	"fmt"
	"net"

	"github.com/digitalocean/go-libvirt"
	"github.com/zakariakebairia/kvmcli/internal/registry"
)

func modifyDHCPHost(
	session registry.Session,
	nw libvirt.Network,
	ip net.IP,
	mac net.HardwareAddr,
	flags libvirt.NetworkUpdateFlags,
) error {
	xmlEntry := dhcpHostXML(mac, ip)

	return session.Conn.NetworkUpdate(
		nw,
		uint32(libvirt.NetworkUpdateCommandModify),
		uint32(libvirt.NetworkSectionIPDhcpHost),
		-1,
		xmlEntry,
		flags,
	)
}

func deleteDHCPHost(
	session registry.Session,
	nw libvirt.Network,
	mac net.HardwareAddr,
	flags libvirt.NetworkUpdateFlags,
) error {
	selector := dhcpHostSelectorXML(mac)

	return session.Conn.NetworkUpdate(
		nw,
		uint32(libvirt.NetworkUpdateCommandDelete),
		uint32(libvirt.NetworkSectionIPDhcpHost),
		-1,
		selector,
		flags,
	)
}

func addDHCPHost(
	session registry.Session,
	nw libvirt.Network,
	ip net.IP,
	mac net.HardwareAddr,
	flags libvirt.NetworkUpdateFlags,
) error {
	xml := dhcpHostXML(mac, ip)

	return session.Conn.NetworkUpdate(
		nw,
		uint32(libvirt.NetworkUpdateCommandAddLast),
		uint32(libvirt.NetworkSectionIPDhcpHost),
		-1,
		xml,
		flags,
	)
}

func dhcpHostXML(mac net.HardwareAddr, ip net.IP) string {
	return fmt.Sprintf(`<host mac='%s' ip='%s'/>`, mac.String(), ip.String())
}

func dhcpHostSelectorXML(mac net.HardwareAddr) string {
	return fmt.Sprintf(`<host mac='%s'/>`, mac.String())
}

// TODO: I need to fix this, not clean, over engineered
// SetStaticMapping ensures a DHCP reservation (MAC → IP) exists on a libvirt network.
func SetStaticMapping(session registry.Session, spec *registry.Object, hostAddr *HostAddr) error {
	networkName := spec.GetString("network")

	flags := libvirt.NetworkUpdateAffectLive | libvirt.NetworkUpdateAffectConfig

	nw, err := session.Conn.NetworkLookupByName(networkName)
	if err != nil {
		return fmt.Errorf("lookup network %q: %w", networkName, err)
	}
	if err := modifyDHCPHost(session, nw, hostAddr.IP, hostAddr.MAC, flags); err == nil {
		return nil
	}

	_ = deleteDHCPHost(session, nw, hostAddr.MAC, flags)

	if err := addDHCPHost(session, nw, hostAddr.IP, hostAddr.MAC, flags); err != nil {
		return fmt.Errorf(
			"set dhcp mapping on network %q (mac=%s ip=%s): %w",
			networkName,
			hostAddr.MAC,
			hostAddr.IP,
			err,
		)
	}

	return nil
}
