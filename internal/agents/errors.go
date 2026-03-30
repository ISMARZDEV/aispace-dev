package agents

import (
	"errors"
	"fmt"

	"github.com/ismartz/aispace-setup/internal/model"
)

var (
	ErrCapabilityNotSupported = errors.New("capability not supported")
	ErrAgentNotSupported      = errors.New("agent not supported")
	ErrDuplicateAdapter       = errors.New("adapter already registered")
)

// CapabilityNotSupportedError is returned when an agent is asked to perform
// an action it does not support (e.g., slash commands on Claude Code).
type CapabilityNotSupportedError struct {
	Agent      model.AgentID
	Capability Capability
}

func (e CapabilityNotSupportedError) Error() string {
	return fmt.Sprintf("agent %q does not support capability %q", e.Agent, e.Capability)
}

func (e CapabilityNotSupportedError) Is(target error) bool {
	return target == ErrCapabilityNotSupported
}

// AgentNotSupportedError is returned by NewAdapter for unknown agent IDs.
type AgentNotSupportedError struct {
	Agent model.AgentID
}

func (e AgentNotSupportedError) Error() string {
	return fmt.Sprintf("agent %q is not supported", e.Agent)
}

func (e AgentNotSupportedError) Is(target error) bool {
	return target == ErrAgentNotSupported
}
