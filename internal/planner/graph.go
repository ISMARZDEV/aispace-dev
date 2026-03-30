package planner

import "github.com/ismartz/aispace-setup/internal/model"

// Graph holds the hard dependency edges between components.
type Graph struct {
	dependencies map[model.ComponentID][]model.ComponentID
}

// NewGraph creates a Graph from a dependency map.
// Dependencies are copied defensively so the caller's map can be modified safely.
func NewGraph(deps map[model.ComponentID][]model.ComponentID) Graph {
	normalized := make(map[model.ComponentID][]model.ComponentID, len(deps))
	for component, d := range deps {
		cp := make([]model.ComponentID, len(d))
		copy(cp, d)
		normalized[component] = cp
	}
	return Graph{dependencies: normalized}
}

// Has reports whether component is a known node in the graph.
func (g Graph) Has(component model.ComponentID) bool {
	_, ok := g.dependencies[component]
	return ok
}

// DependenciesOf returns a defensive copy of component's direct dependencies.
func (g Graph) DependenciesOf(component model.ComponentID) []model.ComponentID {
	deps, ok := g.dependencies[component]
	if !ok {
		return nil
	}
	cp := make([]model.ComponentID, len(deps))
	copy(cp, deps)
	return cp
}

// DefaultGraph returns the canonical component dependency graph for AISpace Setup.
//
// Hard dependency edges:
//   - SDD requires Engram (SDD writes to Engram during phases)
//   - Skills requires SDD (skills reference SDD workflow steps)
//
// All other components are independent leaf nodes.
func DefaultGraph() Graph {
	return NewGraph(map[model.ComponentID][]model.ComponentID{
		model.ComponentEngram:     nil,
		model.ComponentSDD:        {model.ComponentEngram},
		model.ComponentSkills:     {model.ComponentSDD},
		model.ComponentContext7:   nil,
		model.ComponentPersona:    nil,
		model.ComponentPermission: nil,
		model.ComponentTheme:      nil,
		model.ComponentAISpace:    nil,
	})
}

// softOrderingPairs defines pairs where first MUST execute before second when
// BOTH are present. These are NOT hard deps — selecting one does NOT auto-add the other.
//
// INVARIANT: the first element in every pair must have nil deps in DefaultGraph.
// (If it had deps it would already be sorted later, making the soft rule redundant.)
//
// Reason: StrategyFileReplace agents (OpenCode) have Persona write the base file and
// SDD/Engram append to it. If SDD ran first, Persona would overwrite its sections.
var softOrderingPairs = [][2]model.ComponentID{
	{model.ComponentPersona, model.ComponentEngram},
	{model.ComponentPersona, model.ComponentSDD},
}

// SoftOrderingConstraints returns the static soft-ordering pairs.
func SoftOrderingConstraints() [][2]model.ComponentID {
	return softOrderingPairs
}
