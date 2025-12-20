package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// GlobalConfig represents the system-wide configuration loaded from TOML.
type GlobalConfig struct {
	Meta     MetaConfig     `toml:"meta"`
	Paths    PathsConfig    `toml:"paths"`
	VM       VMDefaults     `toml:"vm"`
	Domain   DomainDefaults `toml:"domain"`
	Disk     DiskDefaults   `toml:"disk"`
	Net      NetDefaults    `toml:"net"`
	Graphics GraphicsConfig `toml:"graphics"`
	QEMU     QEMUConfig     `toml:"qemu"`
}

type MetaConfig struct {
	Version int `toml:"version"`
}

type PathsConfig struct {
	DB        string `toml:"db"`
	ImagesDir string `toml:"images_dir"`
}

type VMDefaults struct {
	Defaults VMDefaultSettings `toml:"defaults"`
}

type VMDefaultSettings struct {
	CPU        int    `toml:"cpu"`
	Memory     string `toml:"memory"`
	Disk       string `toml:"disk"`
	NamePrefix string `toml:"name_prefix"`
	Namespace  string `toml:"namespace"`
}

type DomainDefaults struct {
	Defaults DomainDefaultSettings `toml:"defaults"`
}

type DomainDefaultSettings struct {
	Machine    string `toml:"machine"`
	Arch       string `toml:"arch"`
	DomainType string `toml:"domain_type"`
	BootDevice string `toml:"boot_device"`
}

type DiskDefaults struct {
	Defaults DiskDefaultSettings `toml:"defaults"`
}

type DiskDefaultSettings struct {
	Bus          string `toml:"bus"`
	Format       string `toml:"format"`
	TargetPrefix string `toml:"target_prefix"`
}

type NetDefaults struct {
	Defaults NetDefaultSettings `toml:"defaults"`
}

type NetDefaultSettings struct {
	Type  string `toml:"type"`
	Model string `toml:"model"`
}

type GraphicsConfig struct {
	Defaults GraphicsDefaultSettings `toml:"defaults"`
}

type GraphicsDefaultSettings struct {
	Type     string `toml:"type"`
	Listen   string `toml:"listen"`
	Autoport bool   `toml:"autoport"`
}

type QEMUConfig struct {
	MachineAliases map[string]string `toml:"machine_aliases"`
}

// DefaultGlobalConfig returns a configuration with sensible defaults.
func DefaultGlobalConfig() GlobalConfig {
	home, _ := os.UserHomeDir()
	return GlobalConfig{
		Meta: MetaConfig{Version: 1},
		Paths: PathsConfig{
			DB:        filepath.Join(home, ".local", "share", "kvmcli", "kvmcli.db"),
			ImagesDir: filepath.Join(home, ".local", "share", "kvmcli", "images"),
		},
		VM: VMDefaults{
			Defaults: VMDefaultSettings{
				CPU:       2,
				Memory:    "2GiB",
				Disk:      "20GiB",
				Namespace: "default",
			},
		},
		Domain: DomainDefaults{
			Defaults: DomainDefaultSettings{
				Machine:    "q35",
				Arch:       "x86_64",
				DomainType: "kvm",
				BootDevice: "hd",
			},
		},
		Disk: DiskDefaults{
			Defaults: DiskDefaultSettings{
				Bus:          "virtio",
				Format:       "qcow2",
				TargetPrefix: "vd",
			},
		},
		Net: NetDefaults{
			Defaults: NetDefaultSettings{
				Type:  "network",
				Model: "virtio",
			},
		},
		Graphics: GraphicsConfig{
			Defaults: GraphicsDefaultSettings{
				Type:     "vnc",
				Listen:   "0.0.0.0",
				Autoport: true,
			},
		},
		QEMU: QEMUConfig{
			MachineAliases: map[string]string{
				"q35": "pc-q35-9.2",
				"pc":  "pc-i440fx-9.2",
			},
		},
	}
}

// LoadGlobal looks for kvmcli.toml in standard locations and merges them.
// Precedence: explicit path > ./kvmcli.toml > ~/.config/kvmcli/config.toml > /etc/kvmcli/kvmcli.toml > defaults
func LoadGlobal(explicitPath string) (*GlobalConfig, error) {
	cfg := DefaultGlobalConfig()

	paths := []string{}

	// 1. /etc/kvmcli/kvmcli.toml
	paths = append(paths, "/etc/kvmcli/kvmcli.toml")

	// 2. ~/.config/kvmcli/config.toml
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".config", "kvmcli", "config.toml"))
		// Also try ~/.config/kvmcli/kvmcli.toml as usually expected
		paths = append(paths, filepath.Join(home, ".config", "kvmcli", "kvmcli.toml"))
	}

	// 3. ./kvmcli.toml
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(cwd, "kvmcli.toml"))
		// Also try ./configs/kvmcli.toml for local dev convenience
		paths = append(paths, filepath.Join(cwd, "configs", "kvmcli.toml"))
	}

	// 4. Explicit path (highest priority)
	if explicitPath != "" {
		paths = append(paths, explicitPath)
	}

	loadedAny := false
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			content, err := os.ReadFile(p)
			if err != nil {
				return nil, fmt.Errorf("read config %s: %w", p, err)
			}

			// Unmarshal into the existing struct to overwrite defaults
			if err := toml.Unmarshal(content, &cfg); err != nil {
				return nil, fmt.Errorf("parse config %s: %w", p, err)
			}
			loadedAny = true
		}
	}

	if !loadedAny && explicitPath != "" {
		return nil, fmt.Errorf("explicit config path %q not found", explicitPath)
	}

	return &cfg, nil
}
