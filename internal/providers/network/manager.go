package network

import (
	"fmt"

	"github.com/digitalocean/go-libvirt"
	"github.com/zakariakebairia/kvmcli/internal/registry"
)

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
