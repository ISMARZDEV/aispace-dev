package agents

import (
	"fmt"

	"github.com/ismartz/aispace-setup/internal/agents/claude"
	"github.com/ismartz/aispace-setup/internal/agents/opencode"
	"github.com/ismartz/aispace-setup/internal/model"
)

// NewAdapter returns the Adapter for the given agent.
// Returns AgentNotSupportedError for unknown agent IDs.
func NewAdapter(agent model.AgentID) (Adapter, error) {
	switch agent {
	case model.AgentClaudeCode:
		return claude.NewAdapter(), nil
	case model.AgentOpenCode:
		return opencode.NewAdapter(), nil
	default:
		return nil, AgentNotSupportedError{Agent: agent}
	}
}

// NewMVPRegistry creates a Registry with Claude Code and OpenCode adapters.
func NewMVPRegistry() (*Registry, error) {
	claudeAdapter, err := NewAdapter(model.AgentClaudeCode)
	if err != nil {
		return nil, fmt.Errorf("create claude adapter: %w", err)
	}
	opencodeAdapter, err := NewAdapter(model.AgentOpenCode)
	if err != nil {
		return nil, fmt.Errorf("create opencode adapter: %w", err)
	}
	registry, err := NewRegistry(claudeAdapter, opencodeAdapter)
	if err != nil {
		return nil, fmt.Errorf("create registry: %w", err)
	}
	return registry, nil
}
