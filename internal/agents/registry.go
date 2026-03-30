package agents

import (
	"fmt"
	"slices"

	"github.com/ismartz/aispace-setup/internal/model"
)

// Registry holds the set of available agent adapters, keyed by AgentID.
type Registry struct {
	adapters map[model.AgentID]Adapter
}

// NewRegistry creates a Registry from the provided adapters.
// Returns an error if any adapter is nil or if two adapters share the same AgentID.
func NewRegistry(adapters ...Adapter) (*Registry, error) {
	r := &Registry{adapters: make(map[model.AgentID]Adapter, len(adapters))}
	for _, adapter := range adapters {
		if err := r.Register(adapter); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// Register adds an adapter to the registry.
// Returns an error if the adapter is nil or the agent is already registered.
func (r *Registry) Register(adapter Adapter) error {
	if adapter == nil {
		return fmt.Errorf("adapter is nil")
	}
	agent := adapter.Agent()
	if _, exists := r.adapters[agent]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateAdapter, agent)
	}
	r.adapters[agent] = adapter
	return nil
}

// Get returns the adapter for the given agent, or false if not registered.
func (r *Registry) Get(agent model.AgentID) (Adapter, bool) {
	adapter, ok := r.adapters[agent]
	return adapter, ok
}

// SupportedAgents returns all registered agent IDs in sorted order.
func (r *Registry) SupportedAgents() []model.AgentID {
	ids := make([]model.AgentID, 0, len(r.adapters))
	for id := range r.adapters {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids
}
