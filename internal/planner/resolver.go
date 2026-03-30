package planner

import (
	"fmt"

	"github.com/ismartz/aispace-setup/internal/model"
)

// ResolvedPlan is the output of dependency resolution.
type ResolvedPlan struct {
	// OrderedComponents is the full install order, including auto-added dependencies.
	OrderedComponents []model.ComponentID
	// AddedDependencies are components not in the user's selection but required by deps.
	AddedDependencies []model.ComponentID
}

// Resolver resolves a user Selection into a ResolvedPlan.
type Resolver interface {
	Resolve(selection model.Selection) (ResolvedPlan, error)
}

type dependencyResolver struct {
	graph Graph
}

// NewResolver creates a Resolver backed by graph.
func NewResolver(graph Graph) Resolver {
	return dependencyResolver{graph: graph}
}

func (r dependencyResolver) Resolve(selection model.Selection) (ResolvedPlan, error) {
	resolved := ResolvedPlan{}

	selectedSet := make(map[model.ComponentID]struct{}, len(selection.Components))
	dependencies := map[model.ComponentID][]model.ComponentID{}

	for _, selected := range selection.Components {
		if !r.graph.Has(selected) {
			return ResolvedPlan{}, fmt.Errorf("unknown component %q", selected)
		}
		selectedSet[selected] = struct{}{}
		if err := r.expandDependencies(selected, dependencies); err != nil {
			return ResolvedPlan{}, err
		}
	}

	orderedComponents, err := TopologicalSort(dependencies)
	if err != nil {
		return ResolvedPlan{}, err
	}

	orderedComponents = applySoftOrdering(orderedComponents, SoftOrderingConstraints())

	for _, component := range orderedComponents {
		if _, selected := selectedSet[component]; !selected {
			resolved.AddedDependencies = append(resolved.AddedDependencies, component)
		}
	}

	resolved.OrderedComponents = orderedComponents
	return resolved, nil
}

func (r dependencyResolver) expandDependencies(
	component model.ComponentID,
	dependencies map[model.ComponentID][]model.ComponentID,
) error {
	if _, visited := dependencies[component]; visited {
		return nil
	}

	deps := r.graph.DependenciesOf(component)
	dependencies[component] = deps

	for _, dep := range deps {
		if !r.graph.Has(dep) {
			return fmt.Errorf("component %q depends on unknown dependency %q", component, dep)
		}
		if err := r.expandDependencies(dep, dependencies); err != nil {
			return err
		}
	}

	return nil
}
