package network

import (
	"fmt"
	"net"

	"github.com/digitalocean/go-libvirt"
)

func (m *NetworkManager) modifyDHCPHost(
	nw libvirt.Network,
	ip net.IP,
	mac net.HardwareAddr,
	flags libvirt.NetworkUpdateFlags,
) error {
	xmlEntry := dhcpHostXML(mac, ip)

	return m.conn.NetworkUpdate(
		nw,
		uint32(libvirt.NetworkUpdateCommandModify),
		uint32(libvirt.NetworkSectionIPDhcpHost),
		-1,
		xmlEntry,
		flags,
	)
}

func (m *NetworkManager) deleteDHCPHost(
	nw libvirt.Network,
	mac net.HardwareAddr,
	flags libvirt.NetworkUpdateFlags,
) error {
	selector := dhcpHostSelectorXML(mac)

	return m.conn.NetworkUpdate(
		nw,
		uint32(libvirt.NetworkUpdateCommandDelete),
		uint32(libvirt.NetworkSectionIPDhcpHost),
		-1,
		selector,
		flags,
	)
}

func (m *NetworkManager) addDHCPHost(
	nw libvirt.Network,
	ip net.IP,
	mac net.HardwareAddr,
	flags libvirt.NetworkUpdateFlags,
) error {
	xml := dhcpHostXML(mac, ip)

	return m.conn.NetworkUpdate(
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
