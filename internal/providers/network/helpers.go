package network

import (
	"encoding/xml"
	"fmt"
	"net"

	"github.com/zakariakebairia/kvmcli/internal/registry"
	"github.com/zakariakebairia/kvmcli/internal/templates"
)

func MacFromIP(macPrefix, ipStr string) (string, error) {
	ip := net.ParseIP(ipStr).To4()
	if ip == nil {
		return "", fmt.Errorf("invalid IPv4 address %q", ipStr)
	}
	m, err := net.ParseMAC(macPrefix + ":00:00:00")
	if err != nil {
		return "", fmt.Errorf("invalid MAC prefix %q", macPrefix)
	}
	return fmt.Sprintf(
		"%02x:%02x:%02x:00:%02x:%02x",
		m[0], m[1], m[2],
		ip[2], ip[3],
	), nil
}

func ResolveMAC(prefix, ip, specMAC string) (string, error) {
	// Explicit MAC always wins
	if specMAC != "" {
		return specMAC, nil
	}

	// No MAC, no IP → nothing to resolve
	if ip == "" {
		return "", nil
	}

	mac, err := MacFromIP(prefix, ip)
	if err != nil {
		return "", err
	}

	return mac, nil
}

// prepareNetworkXML generates the libvirt XML definition from an Object.
func prepareNetworkXML(obj *registry.Object) (string, error) {
	var opts []templates.NetworkOption

	if dhcp, ok := obj.Attrs["dhcp"].(map[string]any); ok {
		start, _ := dhcp["start"].(string)
		end, _ := dhcp["end"].(string)
		if start != "" && end != "" {
			opts = append(opts, templates.WithDHCP(start, end))
		}
	}

	if bridge := attrStr(*obj, "bridge"); bridge != "" {
		opts = append(opts, templates.WithBridge(bridge))
	}

	netXML := templates.NewNetwork(
		obj.Name,
		attrStr(*obj, "mode"),
		attrStr(*obj, "net_address"),
		attrStr(*obj, "netmask"),
		attrBool(*obj, "autostart"),
		opts...,
	)

	xmlConfig, err := netXML.GenerateXML()
	if err != nil {
		return "", fmt.Errorf("generate XML for network %s: %w", obj.Name, err)
	}

	return xml.Header + string(xmlConfig), nil
}

func attrBool(s registry.Object, key string) bool {
	value, _ := s.Attrs[key].(bool)
	return value
}
