package network

import (
	"database/sql"
	"fmt"
	"net"

	"github.com/digitalocean/go-libvirt"
)

type NetworkManager struct {
	conn *libvirt.Libvirt
	db   *sql.DB
}

func NewNetworkManager(conn *libvirt.Libvirt, db *sql.DB) *NetworkManager {
	return &NetworkManager{conn: conn, db: db}
}

// SetStaticMapping ensures a DHCP reservation (MAC → IP) exists on a libvirt network.
//
// Semantics:
// - If mapping exists: update it to the requested IP (Modify; fallback Delete+Add)
// - If mapping does not exist: create it
func (m *NetworkManager) SetStaticMapping(
	networkName string,
	ip net.IP,
	mac net.HardwareAddr,
) error {
	nw, err := m.conn.NetworkLookupByName(networkName)
	if err != nil {
		return fmt.Errorf("lookup network %q: %w", networkName, err)
	}

	flags := libvirt.NetworkUpdateAffectLive | libvirt.NetworkUpdateAffectConfig

	// 1) Modify (clean path if host entry exists)
	if err := m.modifyDHCPHost(nw, ip, mac, flags); err == nil {
		return nil
	}

	// 2) Fallback: Delete (ignore if missing) then Add
	// Delete by MAC selector only.
	_ = m.deleteDHCPHost(nw, mac, flags)

	if err := m.addDHCPHost(nw, ip, mac, flags); err != nil {
		return fmt.Errorf(
			"set dhcp mapping on network %q (mac=%s ip=%s): %w",
			networkName,
			mac,
			ip,
			err,
		)
	}

	return nil
}
