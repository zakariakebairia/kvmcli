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
