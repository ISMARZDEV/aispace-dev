package planner

import (
	"fmt"

	"github.com/ismartz/aispace-setup/internal/model"
)

// TopologicalSort returns a deterministic install order respecting all dependency edges.
// Returns an error if a cycle is detected.
func TopologicalSort(dependencies map[model.ComponentID][]model.ComponentID) ([]model.ComponentID, error) {
	visited := map[model.ComponentID]bool{}
	inStack := map[model.ComponentID]bool{}
	var order []model.ComponentID

	var visit func(model.ComponentID) error
	visit = func(node model.ComponentID) error {
		if inStack[node] {
			return fmt.Errorf("dependency cycle detected at %q", node)
		}
		if visited[node] {
			return nil
		}

		inStack[node] = true
		for _, dep := range dependencies[node] {
			if err := visit(dep); err != nil {
				return err
			}
		}
		inStack[node] = false
		visited[node] = true
		order = append(order, node)
		return nil
	}

	// Iterate in deterministic order using a sorted key list.
	keys := sortedKeys(dependencies)
	for _, node := range keys {
		if err := visit(node); err != nil {
			return nil, err
		}
	}

	return order, nil
}

// applySoftOrdering reorders components so that for each (first, second) pair,
// if both are present, first appears before second.
// It does NOT add missing components.
func applySoftOrdering(components []model.ComponentID, pairs [][2]model.ComponentID) []model.ComponentID {
	for _, pair := range pairs {
		first, second := pair[0], pair[1]
		firstIdx, secondIdx := -1, -1
		for i, c := range components {
			if c == first {
				firstIdx = i
			}
			if c == second {
				secondIdx = i
			}
		}
		// Both present and out of order → move first to just before second.
		if firstIdx != -1 && secondIdx != -1 && firstIdx > secondIdx {
			reordered := make([]model.ComponentID, 0, len(components))
			for i, c := range components {
				if i == firstIdx {
					continue // skip first in its current position
				}
				if i == secondIdx {
					reordered = append(reordered, first) // insert first before second
				}
				reordered = append(reordered, c)
			}
			components = reordered
		}
	}
	return components
}

// sortedKeys returns the keys of m in a deterministic alphabetical order.
func sortedKeys(m map[model.ComponentID][]model.ComponentID) []model.ComponentID {
	keys := make([]model.ComponentID, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// Simple insertion sort — small N.
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}
