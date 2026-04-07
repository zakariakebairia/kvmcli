package engine

import (
	"slices"

	"github.com/zakariakebairia/kvmcli/internal/registry"
)

func sortByDependency(objects []registry.Object, reverse bool) [][]registry.Object {
	// Step 1: Group objects by their type name
	objectsByType := make(map[string][]registry.Object)
	for _, object := range objects {
		objectsByType[object.TypeName] = append(objectsByType[object.TypeName], object)
	}

	// Step 2: Set up tracking
	remaining := make(map[string]bool)
	for typeName := range objectsByType {
		remaining[typeName] = true
	}
	resolved := make(map[string]bool)

	// Step 3: Build levels by resolving dependencies round by round
	var levels [][]registry.Object

	for len(remaining) > 0 {
		var currentLevel []registry.Object
		var readyThisRound []string

		// Step 4: Find types whose dependencies are all resolved
		for typeName := range remaining {
			resourceType, ok := registry.Get(typeName)
			if !ok {
				readyThisRound = append(readyThisRound, typeName)
				continue
			}

			allDepsResolved := true
			for _, dep := range resourceType.DependsOn {
				if !resolved[dep] {
					allDepsResolved = false
					break
				}
			}

			if allDepsResolved {
				currentLevel = append(currentLevel, objectsByType[typeName]...)
				readyThisRound = append(readyThisRound, typeName)
			}
		}

		// Step 5: Circular dependency safety
		if len(readyThisRound) == 0 {
			break
		}

		// Step 6: Move ready types from remaining to resolved
		for _, typeName := range readyThisRound {
			resolved[typeName] = true
			delete(remaining, typeName)
		}

		if len(currentLevel) > 0 {
			levels = append(levels, currentLevel)
		}
	}

	// Step 7: Reverse for delete operations
	if reverse {
		slices.Reverse(levels)
	}

	return levels
}
