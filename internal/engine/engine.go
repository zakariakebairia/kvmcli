package engine

import (
	"fmt"

	"github.com/zakariakebairia/kvmcli/internal/database"
	"github.com/zakariakebairia/kvmcli/internal/registry"
)

// Engine orchestrates resource lifecycle.
// It looks up the correct provider via the registry, calls its lifecycle methods,
// and persists state via the DBHandler.
type Engine struct {
	dbHandler *database.DBHandler
	ctx       registry.Session
}

func New(ctx registry.Session, dbHandler *database.DBHandler) *Engine {
	return &Engine{dbHandler: dbHandler, ctx: ctx}
}

// Apply creates each desired resource and persists its state.
// Later this will diff desired vs current and handle dependency ordering.
func (e *Engine) Apply(desired []registry.Object) error {
	for _, obj := range desired {
		rt, ok := registry.Get(obj.TypeName)
		if !ok {
			return fmt.Errorf("unknown resource type: %s", obj.TypeName)
		}

		change := registry.Change{
			Action:  registry.ActionCreate,
			Desired: &obj,
			Current: nil,
		}

		if err := rt.Lifecycle.Apply(e.ctx, change); err != nil {
			return fmt.Errorf("apply %s/%s: %w", obj.TypeName, obj.Name, err)
		}

		obj.Status = "created"
		if err := e.dbHandler.Put(e.ctx.Ctx, &obj); err != nil {
			return fmt.Errorf("save state %s/%s: %w", obj.TypeName, obj.Name, err)
		}
	}
	return nil
}

// Destroy tears down each target resource and removes its state.
// Later this will handle reverse dependency ordering.
func (e *Engine) Destroy(targets []registry.Object) error {
	for _, obj := range targets {
		rt, ok := registry.Get(obj.TypeName)
		if !ok {
			fmt.Printf("warning: unknown resource type %s, skipping\n", obj.TypeName)
			continue
		}

		if err := rt.Lifecycle.Destroy(e.ctx, obj); err != nil {
			fmt.Printf("warning: failed to destroy %s/%s: %v\n", obj.TypeName, obj.Name, err)
			continue
		}

		if err := e.dbHandler.Remove(e.ctx.Ctx, obj.TypeName, obj.Name, obj.Namespace); err != nil {
			fmt.Printf("warning: failed to remove state %s/%s: %v\n", obj.TypeName, obj.Name, err)
			continue
		}
	}
	return nil
}
