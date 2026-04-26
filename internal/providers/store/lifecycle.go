package store

import (
	"github.com/zakariakebairia/kvmcli/internal/registry"
)

const (
	storeObjName = "store"
)

func init() {
	registry.Register(&registry.ResourceType{
		Name:      storeObjName,
		DependsOn: []string{}, // stores have no dependencies
		Lifecycle: &StoreLifecycle{},
		Columns:   []string{"NAME", "NAMESPACE", "BACKEND", "ARTIFACTS", "IMAGES", "STATUS"},
		Format: func(s registry.Object) []string {
			return []string{
				s.Name,
				s.Namespace,
				s.GetString("backend"),
				s.GetString("artifacts_path"),
				s.GetString("images_path"),
				s.Status,
			}
		},
	})
}

// I might change this into a Pool and Volumes
// StoreLifecycle implements registry.ResourceLifecycle.
type StoreLifecycle struct{}

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
	return nil
}

func (l *StoreLifecycle) Destroy(session registry.Session, change registry.Change) error {
	// Images are cleaned up by ON DELETE CASCADE in the images table FK.
	// The engine handles removing the state from the resources table.
	return nil
}
