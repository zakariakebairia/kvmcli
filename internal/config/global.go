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
VM       VMConfig       `toml:"vm"`
	Domain   DomainConfig   `toml:"domain"`
	Disk     DiskConfig     `toml:"disk"`
	Network  NetworkConfig  `toml:"network"`
	Graphics GraphicsConfig `toml:"graphics"`
	Aliases  MachineAliases `toml:"machine_aliases"`
}

type MetaConfig struct {
	Version int `toml:"version"`
}

type PathsConfig struct {
	DB        string `toml:"db"`
	ImagesDir string `toml:"images_dir"`
}

type VMConfig struct {
	CPU       int    `toml:"cpu"`
	Memory    string `toml:"memory"`
	Disk      string `toml:"disk"`
	Namespace string `toml:"namespace"`
}

type DomainConfig struct {
	Machine    string `toml:"machine"`
	Arch       string `toml:"arch"`
	Type       string `toml:"type"`
	BootDevice string `toml:"boot_device"`
}

type DiskConfig struct {
	Bus          string `toml:"bus"`
	Format       string `toml:"format"`
	TargetPrefix string `toml:"target_prefix"`
}

type NetworkConfig struct {
	Type  string `toml:"type"`
	Model string `toml:"model"`
}

type GraphicsConfig struct {
	Type     string `toml:"type"`
	Listen   string `toml:"listen"`
	Autoport bool   `toml:"autoport"`
}

type MachineAliases map[string]string

// DefaultGlobalConfig returns a configuration with sensible defaults.
func DefaultGlobalConfig() GlobalConfig {
	home, _ := os.UserHomeDir()
	return GlobalConfig{
		Meta: MetaConfig{Version: 1},
		Paths: PathsConfig{
			DB:        filepath.Join(home, ".local", "share", "kvmcli", "kvmcli.db"),
			ImagesDir: filepath.Join(home, ".local", "share", "kvmcli", "images"),
		},
		VM: VMConfig{
			CPU:       2,
			Memory:    "2GiB",
			Disk:      "20GiB",
			Namespace: "default",
		},
		Domain: DomainConfig{
			Machine:    "q35",
			Arch:       "x86_64",
			Type:       "kvm",
			BootDevice: "hd",
		},
		Disk: DiskConfig{
			Bus:          "virtio",
			Format:       "qcow2",
			TargetPrefix: "vd",
		},
		Network: NetworkConfig{
			Type:  "network",
			Model: "virtio",
		},
		Graphics: GraphicsConfig{
			Type:     "vnc",
			Listen:   "0.0.0.0",
			Autoport: true,
		},
		Aliases: MachineAliases{
			"q35": "pc-q35-9.2",
			"pc":  "pc-i440fx-9.2",
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
