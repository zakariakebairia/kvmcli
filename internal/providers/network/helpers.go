package network

import (
	"encoding/xml"
	"fmt"

	"github.com/zakariakebairia/kvmcli/internal/registry"
	"github.com/zakariakebairia/kvmcli/internal/templates"
)

// prepareNetworkXML generates the libvirt XML definition from an Object.
func buildNetworkXML(obj *registry.Object) (string, error) {
	var opts []templates.NetworkOption

	if dhcp, ok := obj.Attrs["dhcp"].(map[string]any); ok {
		start, _ := dhcp["start"].(string)
		end, _ := dhcp["end"].(string)
		if start != "" && end != "" {
			opts = append(opts, templates.WithDHCP(start, end))
		}
	}

	if bridge := obj.GetString("bridge"); bridge != "" {
		opts = append(opts, templates.WithBridge(bridge))
	}

	netXML := templates.NewNetwork(
		obj.Name,
		obj.GetString("mode"),
		obj.GetString("net_address"),
		obj.GetString("netmask"),
		obj.GetBool("autostart"),
		opts...,
	)

	xmlConfig, err := netXML.GenerateXML()
	if err != nil {
		return "", fmt.Errorf("generate XML for network %s: %w", obj.Name, err)
	}

	return xml.Header + string(xmlConfig), nil
}
