package network

import (
	"encoding/xml"
	"fmt"
	"net"

	"github.com/zakariakebairia/kvmcli/internal/registry"
	"github.com/zakariakebairia/kvmcli/internal/templates"
)

// FIX: I need to simplify and clean this code

func buildNetworkXML(obj *registry.Object) (string, error) {
	var opts []templates.NetworkOption

	dhcpOpt, hasDHCP, err := dhcpOption(obj)
	if err != nil {
		return "", err
	}
	if hasDHCP {
		opts = append(opts, dhcpOpt)
	}

	if bridge := obj.GetString("bridge"); bridge != "" {
		opts = append(opts, templates.WithBridge(bridge))
	}

	networkAddr, networkMask, err := parseCIDR(obj.GetString("cidr"))
	if err != nil {
		return "", err
	}

	netXML := templates.NewNetwork(
		obj.Name,
		obj.GetString("mode"),
		networkAddr,
		networkMask,
		obj.GetBool("autostart"),
		opts...,
	)

	xmlConfig, err := netXML.GenerateXML()
	if err != nil {
		return "", fmt.Errorf("generate XML for network %s: %w", obj.Name, err)
	}

	return xml.Header + string(xmlConfig), nil
}

func dhcpOption(obj *registry.Object) (templates.NetworkOption, bool, error) {
	raw, ok := obj.Attrs["dhcp"]
	if !ok {
		return nil, false, nil
	}

	dhcp, ok := raw.(map[string]any)
	if !ok {
		return nil, false, fmt.Errorf("dhcp must be a map[string]any")
	}

	start, startOK := dhcp["start"].(string)
	end, endOK := dhcp["end"].(string)
	if !startOK || !endOK {
		return nil, false, fmt.Errorf("dhcp.start and dhcp.end must be strings")
	}

	if start == "" || end == "" {
		return nil, false, nil
	}

	return templates.WithDHCP(start, end), true, nil
}

func parseCIDR(cidr string) (string, string, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", "", fmt.Errorf("invalid cidr %q: %w", cidr, err)
	}
	return ipNet.IP.String(), net.IP(ipNet.Mask).String(), nil
}
