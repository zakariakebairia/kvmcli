package store

import (
	"fmt"

	"github.com/zakariakebairia/kvmcli/internal/registry"
)

const TypeName = "store"

// IDEA: Adding `STATUS` later
var storeColumns = []string{"NAME", "NAMESPACE", "BACKEND", "ARTIFACTS", "IMAGES"}

func init() {
	registry.Register(&registry.ResourceType{
		Name:      TypeName,
		DependsOn: []string{}, // stores have no dependencies
		Lifecycle: &StoreLifecycle{},
		Columns:   storeColumns,
		Format:    formatStore,
	})
}

// I might change this into a Pool and Volumes
// StoreLifecycle implements registry.ResourceLifecycle.
type StoreLifecycle struct{}

func formatStore(obj registry.Object) []string {
	return []string{
		obj.Name,
		obj.Namespace,
		obj.GetString("backend"),
		obj.GetString("artifacts_path"),
		obj.GetString("images_path"),
		// s.Status,
	}
}

func (l *StoreLifecycle) Plan(desired, current *registry.Object) (registry.Action, error) {
	if current == nil && desired != nil {
		return registry.ActionCreate, nil
	}
	if current != nil && desired == nil {
		return registry.ActionDelete, nil
	}
	// Could add update detection here later
	return registry.ActionNone, nil
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
