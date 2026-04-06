package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// GlobalConfig represents the system-wide configuration loaded from TOML.
type GlobalConfig struct {
	Meta     MetaConfig          `toml:"meta"`
	Paths    PathsConfig         `toml:"paths"`
	VM       GlobalVMConfig      `toml:"vm"`
	Domain   DomainConfig        `toml:"domain"`
	Disk     DiskConfig          `toml:"disk"`
	Network  GlobalNetworkConfig `toml:"network"`
	Graphics GraphicsConfig      `toml:"graphics"`
	Aliases  MachineAliases      `toml:"machine_aliases"`
}

type MetaConfig struct {
	Version int `toml:"version"`
}

type PathsConfig struct {
	DB        string `toml:"db"`
	ImagesDir string `toml:"images_dir"`
}

type GlobalVMConfig struct {
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

type GlobalNetworkConfig struct {
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
		VM: GlobalVMConfig{
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
		Network: GlobalNetworkConfig{
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

// LoadDefaultConfig loads the global configuration by searching a set of
// well-known paths in descending priority order:
//
//  1. explicitPath (if provided via --config flag)
//  2. $PWD/kvmcli.toml (project-local)
//  3. ~/.config/kvmcli/kvmcli.toml (user-level)
//  4. /etc/kvmcli/kvmcli.toml (system-wide)
//
// The first file found wins. If no config file is found, an error is returned.
func LoadDefaultConfig() (*GlobalConfig, error) {
	// User home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine home directory: %w", err)
	}
	// Current directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("could not determine current directory: %w", err)
	}
	// respect $XDG_CONFIG_HOME
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig == "" {
		xdgConfig = filepath.Join(home, ".config")
	}

	paths := []string{
		// explicitPath,
		filepath.Join(cwd, "kvmcli.toml"),
		filepath.Join(cwd, "configs", "kvmcli.toml"),
		filepath.Join(home, xdgConfig, "kvmcli", "kvmcli.toml"),
		"/etc/kvmcli/kvmcli.toml",
	}

	for _, p := range paths {
		if p == "" {
			continue
		}
		content, err := os.ReadFile(p)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("read config %q: %w", p, err)
		}

		var cfg GlobalConfig
		if err := toml.Unmarshal(content, &cfg); err != nil {
			return nil, fmt.Errorf("parse config %q: %w", p, err)
		}
		return &cfg, nil
	}

	return nil, fmt.Errorf("no config file found, searched: %s", strings.Join(paths[1:], ", "))
}
