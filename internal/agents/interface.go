package agents

import (
	"context"

	"github.com/ismartz/aispace-setup/internal/model"
	"github.com/ismartz/aispace-setup/internal/system"
)

// Capability tags optional features that an adapter may or may not support.
type Capability string

const (
	CapabilityAutoInstall    Capability = "auto-install"
	CapabilityOutputStyles   Capability = "output-styles"
	CapabilitySlashCommands  Capability = "slash-commands"
)

// Adapter is the core abstraction over an AI agent integration.
// Components use adapter methods instead of switch statements on AgentID,
// making it trivial to support new agents without modifying component logic.
type Adapter interface {
	// Identity
	Agent() model.AgentID
	Tier() model.SupportTier

	// Detection — checks whether the agent binary and config directory exist.
	// Returns: installed (binary found), binaryPath, configPath, configFound (dir exists), err.
	Detect(ctx context.Context, homeDir string) (installed bool, binaryPath string, configPath string, configFound bool, err error)

	// Installation
	SupportsAutoInstall() bool
	InstallCommand(profile system.PlatformProfile) ([][]string, error)

	// Config paths — WHERE to write content for each concern.
	GlobalConfigDir(homeDir string) string
	SystemPromptDir(homeDir string) string
	SystemPromptFile(homeDir string) string
	SkillsDir(homeDir string) string
	SettingsPath(homeDir string) string

	// Config strategies — HOW to inject content (separate from WHERE).
	SystemPromptStrategy() model.SystemPromptStrategy
	MCPStrategy() model.MCPStrategy

	// MCP path resolution
	MCPConfigPath(homeDir string, serverName string) string

	// Optional capabilities
	SupportsOutputStyles() bool
	OutputStyleDir(homeDir string) string

	SupportsSlashCommands() bool
	CommandsDir(homeDir string) string

	SupportsSkills() bool
	SupportsSystemPrompt() bool
	SupportsMCP() bool
}
