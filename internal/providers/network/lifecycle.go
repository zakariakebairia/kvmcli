package network

import (
	"fmt"

	"github.com/zakariakebairia/kvmcli/internal/registry"
)

func init() {
	registry.Register(&registry.ResourceType{
		Name:      "network",
		DependsOn: []string{},
		Lifecycle: &NetworkLifecycle{},
		Columns:   []string{"NAME", "NAMESPACE", "BRIDGE", "MODE", "ADDRESS", "STATUS"},
		Format: func(s registry.Object) []string {
			return []string{
				s.Name,
				s.Namespace,
				attrStr(s, "bridge"),
				attrStr(s, "mode"),
				attrStr(s, "net_address"),
				s.Status,
			}
		},
	})
}

// NetworkLifecycle implements registry.ResourceLifecycle.
type NetworkLifecycle struct{}

func (l *NetworkLifecycle) Plan(desired, current *registry.Object) (registry.Action, error) {
	if current == nil && desired != nil {
		return registry.ActionCreate, nil
	}
	if current != nil && desired == nil {
		return registry.ActionDelete, nil
	}
	return registry.ActionNone, nil
}

func (l *NetworkLifecycle) Apply(session registry.Session, change registry.Change) error {
	spec := change.Desired

	xmlConfig, err := prepareNetworkXML(spec)
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

	// Enable autostart if requested
	if attrBool(*spec, "autostart") {
		if err := session.Conn.NetworkSetAutostart(netInstance, 1); err != nil {
			return fmt.Errorf("set autostart for network %q: %w", spec.Name, err)
		}
	}

	spec.Status = "created"
	fmt.Printf("network/%s created\n", spec.Name)
	return nil
}

func (l *NetworkLifecycle) Destroy(session registry.Session, current registry.Object) error {
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

	fmt.Printf("network/%s deleted\n", current.Name)
	return nil
}

// attrStr extracts a string attribute from an Object.
func attrStr(s registry.Object, key string) string {
	if value, ok := s.Attrs[key].(string); ok {
		return value
	}
	return ""
}
