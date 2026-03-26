package config

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zakariakebairia/kvmcli/internal/database"
	"github.com/zakariakebairia/kvmcli/internal/registry"
)

// Load parses an HCL file, resolves all expressions, and returns
// the resulting Objects ready for the engine.
func Load(path string, ctx context.Context, dbHandler *database.DBHandler) ([]registry.Object, error) {
	cfg, err := parse(path)
	if err != nil {
		return nil, err
	}

	if err := resolve(cfg, ctx, dbHandler); err != nil {
		return nil, err
	}

	return buildObjects(cfg), nil
}

// --- HCL parsing ---------------------------------------------------------

func parse(path string) (*hclConfig, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(src, path)
	if diags.HasErrors() {
		return nil, fmt.Errorf("parse hcl %q: %w", path, diags)
	}

	var cfg hclConfig
	if diags := gohcl.DecodeBody(file.Body, nil, &cfg); diags.HasErrors() {
		return nil, fmt.Errorf("decode hcl %q: %w", path, diags)
	}
	return &cfg, nil
}

// --- HCL config types ----------------------------------------------------

// hclConfig represents a complete kvmcli HCL file.
type hclConfig struct {
	Locals   *hclLocals    `hcl:"locals,block"`
	Networks []networkDef  `hcl:"network,block"`
	VMs      []vmDef       `hcl:"vm,block"`
	Stores   []storeDef    `hcl:"store,block"`
	Data     []dataRef     `hcl:"data,block"`
}

type hclLocals struct {
	Values map[string]hcl.Expression `hcl:",remain"`
}

// dataRef is a reference to a resource that already exists in the DB.
// Example: data "store" "homelab" {}
type dataRef struct {
	Type string `hcl:"type,label"`
	Name string `hcl:"name,label"`
}

// vmDef describes a virtual machine block in HCL.
type vmDef struct {
	Name      string            `hcl:"name,label"`
	Namespace string            `hcl:"namespace"`
	Image     string            `hcl:"image"`
	CPU       int               `hcl:"cpu"`
	Memory    int               `hcl:"memory"`
	Disk      string            `hcl:"disk,optional"`
	NetExpr   hcl.Expression    `hcl:"network,attr"`
	NetName   string
	StoreExpr hcl.Expression    `hcl:"store,attr"`
	Store     string
	MAC       string            `hcl:"mac,optional"`
	IP        string            `hcl:"ip,optional"`
	Labels    map[string]string `hcl:"labels,optional"`
}

// networkDef describes a network block in HCL.
type networkDef struct {
	Name       string            `hcl:"name,label"`
	Namespace  string            `hcl:"namespace"`
	CIDR       string            `hcl:"cidr,optional"`
	NetAddress string            `hcl:"netaddress,optional"`
	NetMask    string            `hcl:"netmask,optional"`
	Bridge     string            `hcl:"bridge,optional"`
	Mode       string            `hcl:"mode,optional"`
	DHCP       *dhcpDef          `hcl:"dhcp,block"`
	Autostart  bool              `hcl:"autostart,optional"`
	Labels     map[string]string `hcl:"labels,optional"`
}

type dhcpDef struct {
	Start string `hcl:"start"`
	End   string `hcl:"end"`
}

// storeDef describes a store block in HCL.
type storeDef struct {
	Name      string            `hcl:"name,label"`
	Namespace string            `hcl:"namespace"`
	Labels    map[string]string `hcl:"labels,optional"`
	Backend   string            `hcl:"backend,optional"`
	Paths     storePathsDef     `hcl:"paths,block"`
	Images    []imageDef        `hcl:"image,block"`
}

type storePathsDef struct {
	Artifacts string `hcl:"artifacts,optional"`
	Images    string `hcl:"images,optional"`
}

type imageDef struct {
	Name      string `hcl:"name,label"`
	Display   string `hcl:"display,optional"`
	Version   string `hcl:"version,optional"`
	OSProfile string `hcl:"os_profile,optional"`
	File      string `hcl:"file,optional"`
	Size      string `hcl:"size,optional"`
	Checksum  string `hcl:"checksum,optional"`
}
