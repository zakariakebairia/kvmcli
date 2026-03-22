package network

import (
	"encoding/xml"
	"fmt"
	"time"

	db "github.com/zakariakebairia/kvmcli/internal/database"
	templates "github.com/zakariakebairia/kvmcli/internal/templates"
)

// prepareNetwork generates the libvirt-compatible XML configuration
// for a virtual network, using optional parameters like DHCP and bridge.
// prepareNetwork generates the libvirt-compatible XML configuration
// for a virtual network, using optional parameters like DHCP and bridge.
func (m *LibvirtNetworkManager) prepareNetwork(spec Config) (string, error) {
	var opts []templates.NetworkOption

	// Append DHCP config if defined in the YAML
	// Append DHCP config if defined in the YAML
	if spec.DHCP != nil {
		opts = append(opts, templates.WithDHCP(spec.DHCP.Start, spec.DHCP.End))
	}

	// Append bridge name if provided
	if spec.Bridge != "" {
		opts = append(opts, templates.WithBridge(spec.Bridge))
	}

	// Create the network definition with all options
	netXML := templates.NewNetwork(
		spec.Name,
		spec.Mode,
		spec.NetAddress,
		spec.NetMask,
		spec.Autostart,
		opts...,
	)

	xmlConfig, err := netXML.GenerateXML()
	if err != nil {
		return "", fmt.Errorf(
			"failed to generate XML for network %s: %v",
			spec.Name,
			err,
		)
	}

	return xml.Header + string(xmlConfig), nil
}

// defineAndStartNetwork defines and starts the virtual network using libvirt.
// defineAndStartNetwork defines and starts the virtual network using libvirt.
func (m *LibvirtNetworkManager) defineAndStartNetwork(xmlConfig string) error {
	// Define the network from the generated XML
	netInstance, err := m.conn.NetworkDefineXML(xmlConfig)
	if err != nil {
		return fmt.Errorf("failed to define network: %v", err)
	}

	// Start (create) the defined network
	if err := m.conn.NetworkCreate(netInstance); err != nil {
		return fmt.Errorf("failed to start network: %w", err)
	}
	// TODO: Autostart handling needs the spec or to be passed in.
	// For now omitting autostart logic here or assuming it's done elsewhere?
	// The original code used `net.Spec.Autostart`.
	// Let's rely on the caller to handle autostart or pass it in.
	// Actually, create.go logic was: prepare -> defineAndStart.
	// I should update defineAndStartNetwork to take the autostart flag or name to lookup spec?
	// Or simpler: just take the flag.

	return nil
}

func (m *LibvirtNetworkManager) EnableAutostart(name string) error {
	netInstance, err := m.conn.NetworkLookupByName(name)
	if err != nil {
		return err
	}
	return m.conn.NetworkSetAutostart(netInstance, 1)
}

func NewNetworkRecord(net *Network) *db.VirtualNetwork {
	// Create a map[string]string for the DB record if DHCP is present
	var dhcpMap map[string]string
	if net.Spec.DHCP != nil {
		dhcpMap = map[string]string{
			"start": net.Spec.DHCP.Start,
			"end":   net.Spec.DHCP.End,
		}
	}

	// Create network record out of infos
	return &db.VirtualNetwork{
		Name:      net.Spec.Name,
		Namespace: net.Spec.Namespace,
		Labels:    net.Spec.Labels,
		// MacAddress: net.Spec.MacAddress,
		Bridge:     net.Spec.Bridge,
		Mode:       net.Spec.Mode,
		NetAddress: net.Spec.NetAddress,
		Netmask:    net.Spec.NetMask,
		DHCP:       dhcpMap,
		Autostart:  net.Spec.Autostart,
		CreatedAt:  time.Now(),
	}
}
