package store

import (
	"fmt"

	"github.com/zakariakebairia/kvmcli/internal/registry"
)

const TypeName = "store"

// IDEA: Adding `STATUS` later
var Columns = []string{
	"NAME",
	"NAMESPACE",
	"BACKEND",
	"ARTIFACTS PATH",
	"IMAGES PATH",
}

func init() {
	registry.Register(&registry.ResourceType{
		Name:      TypeName,
		DependsOn: []string{}, // stores have no dependencies
		Lifecycle: &StoreLifecycle{},
		Columns:   Columns,
		Format:    formatStore,
	})
}

// TODO: Consider splitting the store abstraction into pools and volumes.

// Lifecycle implements the registry.ObjectLifecycle interface
// for store resources.
type StoreLifecycle struct{}

// formatStore renders a store resource for table output.
func formatStore(obj registry.Object) []string {
	return []string{
		obj.Name,
		obj.Namespace,
		obj.GetString("backend"),
		obj.GetString("artifacts_path"),
		obj.GetString("images_path"),
	}
}

func (l *StoreLifecycle) Plan(desired, current *registry.Object) (registry.Action, error) {
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

func (l *StoreLifecycle) Apply(session registry.Session, change registry.Change) error {
	spec := change.Desired
	// check attributes validity
	var attrs StoreAttrs

	if err := attrs.FromObject(spec); err != nil {
		return fmt.Errorf("parsing store %q: %w", spec.Name, err)
	}
	if err := attrs.Validate(); err != nil {
		return err
	}

	return nil
}

func (l *StoreLifecycle) Destroy(session registry.Session, change registry.Change) error {
	// Images are cleaned up by ON DELETE CASCADE in the images table FK.
	// The engine handles removing the state from the resources table.
	return nil
}
