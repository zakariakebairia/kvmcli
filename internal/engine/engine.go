package engine

import (
	"fmt"

	"github.com/zakariakebairia/kvmcli/internal/database"
	"github.com/zakariakebairia/kvmcli/internal/registry"
)

// Engine orchestrates resource lifecycle.
// It looks up the correct provider via the registry, calls its lifecycle methods,
// and persists state via the DBHandler.

// session carries (ctx, sql connection, libvirt connection)

type Engine struct {
	dbHandler *database.DBHandler
	session   registry.Session
}

func New(session registry.Session, dbHandler *database.DBHandler) *Engine {
	return &Engine{dbHandler: dbHandler, session: session}
}

// Apply creates each desired resource and persists its state.
// Later this will diff desired vs current and handle dependency ordering.
func (e *Engine) Apply(desired []registry.Object) error {
	levels := sortByDependency(desired, false)

	for _, level := range levels {
		for _, obj := range level {
			resourceType, ok := registry.Get(obj.TypeName)
			if !ok {
				return fmt.Errorf("unknown resource type: %s", obj.TypeName)
			}

			change := registry.Change{
				Action:  registry.ActionCreate,
				Desired: &obj,
				Current: nil,
			}

			if err := resourceType.Lifecycle.Apply(e.session, change); err != nil {
				return fmt.Errorf("apply %s/%s: %w", obj.TypeName, obj.Name, err)
			}

			obj.Status = "created"
			if err := e.dbHandler.Put(e.session.Ctx, &obj); err != nil {
				return fmt.Errorf("save state %s/%s: %w", obj.TypeName, obj.Name, err)
			}
		}
	}
	return nil
}

// Destroy tears down each target resource and removes its state.
// Later this will handle reverse dependency ordering.
func (e *Engine) Destroy(targets []registry.Object) error {
	levels := sortByDependency(targets, true)

	for _, level := range levels {
		for _, obj := range level {
			rt, ok := registry.Get(obj.TypeName)
			if !ok {
				fmt.Printf("warning: unknown resource type %s, skipping\n", obj.TypeName)
				continue
			}

			if err := rt.Lifecycle.Destroy(e.session, obj); err != nil {
				fmt.Printf("warning: failed to destroy %s/%s: %v\n", obj.TypeName, obj.Name, err)
				continue
			}

			if err := e.dbHandler.Remove(
				e.session.Ctx,
				obj.TypeName,
				obj.Name,
				obj.Namespace,
			); err != nil {
				fmt.Printf("warning: failed to remove state %s/%s: %v\n", obj.TypeName, obj.Name, err)
				continue
			}
		}
	}
	return nil
}
