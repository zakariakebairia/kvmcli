package network

import (
	"fmt"

	"github.com/zakariakebairia/kvmcli/internal/registry"
)

const TypeName = "network"

var Columns = []string{
	"NAME",
	"NAMESPACE",
	"BRIDGE",
	"MODE",
	"ADDRESS",
	"STATUS",
}

func init() {
	registry.Register(&registry.ResourceType{
		Name:      TypeName,
		DependsOn: []string{},
		Lifecycle: &NetworkLifecycle{},
		Columns:   Columns,
		Format:    formatNetwork,
	})
}

// NetworkLifecycle implements registry.ObjectLifecycle.
type NetworkLifecycle struct{}

// formatVM
func formatNetwork(obj registry.Object) []string {
	return []string{
		obj.Name,
		obj.Namespace,
		obj.GetString("bridge"),
		obj.GetString("mode"),
		obj.GetString("cidr"),
		obj.Status,
	}
}

func (l *NetworkLifecycle) Plan(desired, current *registry.Object) (registry.Action, error) {
	switch {
	case current == nil && desired != nil:
		return registry.ActionCreate, nil

	case current != nil && desired == nil:
		return registry.ActionDelete, nil

	case current != nil && desired != nil:
		return registry.ActionUpdate, nil

	default:
		return registry.ActionNone, nil
	}
}

func (l *NetworkLifecycle) Apply(session registry.Session, change registry.Change) error {
	spec := change.Desired
	var attrs NetworkAttrs

	if err := attrs.FromObject(spec); err != nil {
		return fmt.Errorf("parsing net %q: %w", spec.Name, err)
	}
	if err := attrs.Validate(); err != nil {
		return err
	}

	xmlConfig, err := buildNetworkXML(spec)
	if err != nil {
		return err
	}

	// Define the network in libvirt
	netInstance, err := session.Conn.NetworkDefineXML(xmlConfig)
	if err != nil {
		return fmt.Errorf("define network %q: %w", spec.Name, err)
	}

	// Start the network
	if err := session.Conn.NetworkCreate(netInstance); err != nil {
		return fmt.Errorf("start network %q: %w", spec.Name, err)
	}

	if spec.GetBool("autostart") {
		{
			if err := session.Conn.NetworkSetAutostart(netInstance, 1); err != nil {
				return fmt.Errorf("set autostart for network %q: %w", spec.Name, err)
			}
		}
	}

	spec.Status = "created"
	return nil
}

func (l *NetworkLifecycle) Destroy(session registry.Session, change registry.Change) error {
	current := change.Current

	netInstance, err := session.Conn.NetworkLookupByName(current.Name)
	if err != nil {
		return fmt.Errorf("network %q not found: %w", current.Name, err)
	}

	if err := session.Conn.NetworkDestroy(netInstance); err != nil {
		return fmt.Errorf("destroy network %q: %w", current.Name, err)
	}

	if err := session.Conn.NetworkUndefine(netInstance); err != nil {
		return fmt.Errorf("undefine network %q: %w", current.Name, err)
	}

	return nil
}
