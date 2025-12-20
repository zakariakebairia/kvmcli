package config

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/kebairia/kvmcli/internal/network"
	"github.com/kebairia/kvmcli/internal/store"
	"github.com/kebairia/kvmcli/internal/vms"
)

// Config represents a complete kvmcli configuration file.
type Config struct {
	Variables []Variable       `hcl:"variable,block"`
	Locals    *Locals          `hcl:"locals,block"`
	Networks  []network.Config `hcl:"network,block"`
	VMs       []vms.Config     `hcl:"vm,block"`
	Stores    []store.Config   `hcl:"store,block"`
	Clusters  []Cluster        `hcl:"cluster,block"`
	Data      []DataResource   `hcl:"data,block"`
}

// Variable represents a user-defined variable.
type Variable struct {
	Name    string         `hcl:"name,label"`
	Default hcl.Expression `hcl:"default,optional"`
}

// Locals represents a block of local values.
type Locals struct {
	Values map[string]hcl.Expression `hcl:",remain"`
}

type DataResource struct {
	Type string `hcl:"type,label"`
	Name string `hcl:"name,label"`
}

// Cluster describes a logical grouping of VMs.
type Cluster struct {
	Name      string            `hcl:"name,label"`
	VMExprs   hcl.Expression    `hcl:"vms,attr"` // List of VM references
	VMNames   []string          // Resolved VM names
	Labels    map[string]string `hcl:"labels,optional"`
	Lifecycle *Lifecycle        `hcl:"lifecycle,block"`
}

type Lifecycle struct {
	StartOrder []string `hcl:"start_order,optional"`
	StopOrder  []string `hcl:"stop_order,optional"`
}
