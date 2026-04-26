package registry

import (
	"fmt"
	"sync"
)

type ResourceType struct {
	Name      string
	DependsOn []string
	Lifecycle ObjectLifecycle
	Columns   []string
	Format    func(Object) []string
}

// TODO: will be changed later to "ObjectLifeCycle"
type ObjectLifecycle interface {
	Plan(desired, current *Object) (Action, error)
	Apply(session Session, change Change) error
	Destroy(session Session, change Change) error
}

var (
	mu    sync.RWMutex
	types = make(map[string]*ResourceType)
)

func Register(rt *ResourceType) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := types[rt.Name]; exists {
		panic(fmt.Sprintf("resource type %q already registered", rt.Name))
	}
	// TODO: I need to check this, a better way of handling it
	types[rt.Name] = rt
}

// TODO: I will change the return to (*ResourceType, error)
func Get(name string) (*ResourceType, bool) {
	mu.RLock()
	defer mu.RUnlock()
	rt, ok := types[name]
	return rt, ok
}

func All() map[string]*ResourceType {
	mu.RLock()
	defer mu.RUnlock()
	result := make(map[string]*ResourceType, len(types))
	for key, value := range types {
		result[key] = value
	}
	return result
}
