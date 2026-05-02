package network

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/zakariakebairia/kvmcli/internal/registry"
)

// type Mode string

var (
	ModeNAT      = "nat"
	ModeBridge   = "bridge"
	ModeHostOnly = "host-only"
)

// FIX: what is the difference between `mode` and `bridge`
type NetworkAttrs struct {
	AutoStart bool   `json:"auto_start"`
	Bridge    string `json:"bridge"`
	CIDR      string `json:"cidr"`
	Mode      string `json:"mode"` // e.g. bridge, nat, host-only
	// NetworkMask    string `json:"netmask"`
	// NetworkAddress string `json:"netaddress"`
}

func (n *NetworkAttrs) FromObject(object *registry.Object) error {
	data, err := json.Marshal(object.Attrs)
	if err != nil {
		return fmt.Errorf("failed to marshal attrs: %w", err)
	}
	if err := json.Unmarshal(data, n); err != nil {
		return fmt.Errorf("failed to unmarshal attrs: %w", err)
	}
	return nil
}

func (n *NetworkAttrs) Validate() error {
	// Normalize inputs
	mode := strings.ToLower(strings.TrimSpace(n.Mode))
	bridge := strings.TrimSpace(n.Bridge)

	switch mode {
	case ModeNAT, ModeBridge, ModeHostOnly:
	// valid
	default:
		return fmt.Errorf(
			"net.mode must be one of [%s, %s, %s], got: %q",
			ModeNAT, ModeBridge, ModeHostOnly, n.Mode,
		)

	}

	// CIDR validation
	if n.CIDR == "" {
		return fmt.Errorf("net.cidr is required")
	}
	_, _, err := net.ParseCIDR(n.CIDR)
	if err != nil {
		return fmt.Errorf("invalid net.cidr: %w", err)
	}
	// Bridge validation
	switch mode {
	case ModeNAT, ModeBridge:
		if bridge == "" {
			return fmt.Errorf("net.bridge is required when mode=(bridge|nat)")
		}
	case ModeHostOnly:
		if bridge != "" {
			return fmt.Errorf("net.bridge must not be set when mode=%s", mode)
		}
	}

	return nil
}
