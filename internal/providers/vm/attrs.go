package vm

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/zakariakebairia/kvmcli/internal/registry"
)

type VMAttrs struct {
	Image      string `json:"image"`
	Namespace  string `json:"namespace"`
	CPU        int    `json:"cpu"`
	Memory     int    `json:"memory"`
	Disk       string `json:"disk"`
	IP         string `json:"ip"`
	MACAddress string `json:"mac"`
}

func (v *VMAttrs) FromObject(object *registry.Object) error {
	data, err := json.Marshal(object.Attrs)
	if err != nil {
		return fmt.Errorf("failed to marshal attrs: %w", err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal into VMAttrs: %w", err)
	}
	return nil
}

func (v *VMAttrs) Validate() error {
	// --- Required string fields ---
	if v.Image == "" {
		return fmt.Errorf("vm.image is required")
	}
	if v.Disk == "" {
		return fmt.Errorf("vm.disk is required")
	}

	// --- CPU ---
	if v.CPU <= 0 {
		return fmt.Errorf("vm.cpu must be > 0, got %d", v.CPU)
	}
	if v.CPU > 128 {
		return fmt.Errorf("vm.cpu exceeds maximum (128), got %d", v.CPU)
	}

	// --- Memory (in MB) ---
	if v.Memory < 512 {
		return fmt.Errorf("vm.memory must be >= 512 MB, got %d", v.Memory)
	}

	// --- IP ---
	if v.IP == "" {
		return fmt.Errorf("vm.ip is required")
	}
	if net.ParseIP(v.IP) == nil {
		return fmt.Errorf("vm.ip is not a valid IP address: %q", v.IP)
	}

	// --- MAC ---
	// if v.MACAddress == "" {
	// 	return fmt.Errorf("vm.mac is required")
	// }
	// if _, err := net.ParseMAC(v.MACAddress); err != nil {
	// 	return fmt.Errorf("vm.mac is not a valid MAC address: %q", v.MACAddress)
	// }

	return nil
}
