package engine

import (
	"fmt"

	"github.com/zakariakebairia/kvmcli/internal/database"
	logger "github.com/zakariakebairia/kvmcli/internal/logger"
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
			objectType, ok := registry.Get(obj.TypeName)
			if !ok {
				return fmt.Errorf("unknown object type: %s", obj.TypeName)
			}

			change := registry.Change{
				Action:  registry.ActionCreate,
				Desired: &obj,
				Current: nil,
			}

			resource := obj.TypeName + "/" + obj.Name
			if err := objectType.Lifecycle.Apply(e.session, change); err != nil {
				logger.Info(resource, "create", err)
				return fmt.Errorf("apply %s: %w", resource, err)
			}

			obj.Status = "created"
			if err := e.dbHandler.Put(e.session.Ctx, &obj); err != nil {
				return fmt.Errorf("save object %s: %w", resource, err)
			}
			logger.Info(resource, obj.Status, nil)
		}
	}
	return nil
}

// Destroy tears down each target resource and removes its state.
// Later this will handle reverse dependency ordering.
func (e *Engine) Destroy(targets []registry.Object) error {
	levels := sortByDependency(targets, true)

	for _, level := range levels {
		for _, instance := range level {
			resource := instance.TypeName + "/" + instance.Name
			rt, ok := registry.Get(instance.TypeName)
			if !ok {
				logger.Warnf("unknown object type %s, skipping", instance.TypeName)
				continue
			}
			// Get full object info from the database
			object, err := e.dbHandler.Get(
				e.session.Ctx,
				instance.TypeName,
				instance.Name,
				instance.Namespace,
			)
			if err != nil {
				return fmt.Errorf("get %s: %w", object, err)
			}
			if object == nil {
				logger.Errorf("%s not found in database, skipping", resource)
				continue
			}

			change := registry.Change{
				Action:  registry.ActionDelete,
				Desired: nil,
				Current: object,
			}

			if err := rt.Lifecycle.Destroy(e.session, change); err != nil {
				logger.Info(resource, "destroy", err)
				continue
			}

			if err := e.dbHandler.Remove(
				e.session.Ctx,
				object.TypeName,
				object.Name,
				object.Namespace,
			); err != nil {
				// logger.Info(resource, "remove state", err)
				fmt.Printf("%s/%s, remove state, %v", object.TypeName, object.Name, err)
				continue
			}
			fmt.Printf("%s/%s deleted\n", object.TypeName, object.Name)
		}
	}
	return nil
}
