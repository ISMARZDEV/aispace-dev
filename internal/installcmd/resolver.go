package installcmd

import (
	"fmt"

	"github.com/ismartz/aispace-setup/internal/model"
	"github.com/ismartz/aispace-setup/internal/system"
)

// CommandSequence is an ordered list of commands to run in sequence.
// Each inner slice is one command with its arguments.
// Multi-step installs (e.g., brew tap + brew install) use multiple entries.
type CommandSequence = [][]string

// Resolver maps a platform profile to the correct install command sequence.
type Resolver interface {
	ResolveAgentInstall(profile system.PlatformProfile, agent model.AgentID) (CommandSequence, error)
	ResolveComponentInstall(profile system.PlatformProfile, component model.ComponentID) (CommandSequence, error)
}

type profileResolver struct{}

// NewResolver returns the default platform-aware Resolver.
func NewResolver() Resolver {
	return profileResolver{}
}

// ResolveAgentInstall returns the install commands for the given agent on the given platform.
func (profileResolver) ResolveAgentInstall(profile system.PlatformProfile, agent model.AgentID) (CommandSequence, error) {
	switch agent {
	case model.AgentClaudeCode:
		return resolveClaudeCodeInstall(profile), nil
	case model.AgentOpenCode:
		return resolveOpenCodeInstall(profile)
	default:
		return nil, fmt.Errorf("install command not supported for agent %q", agent)
	}
}

// ResolveComponentInstall returns the install commands for the given component.
func (profileResolver) ResolveComponentInstall(profile system.PlatformProfile, component model.ComponentID) (CommandSequence, error) {
	switch component {
	case model.ComponentEngram:
		return resolveEngramInstall(profile)
	default:
		return nil, fmt.Errorf("install command not supported for component %q", component)
	}
}

// resolveClaudeCodeInstall returns the npm install command for Claude Code.
// On Linux without user-writable npm, sudo is required.
func resolveClaudeCodeInstall(profile system.PlatformProfile) CommandSequence {
	if profile.OS == "linux" && !profile.NpmWritable {
		return CommandSequence{{"sudo", "npm", "install", "-g", "@anthropic-ai/claude-code"}}
	}
	return CommandSequence{{"npm", "install", "-g", "@anthropic-ai/claude-code"}}
}

// resolveOpenCodeInstall returns the install commands for OpenCode.
// macOS: brew install via anomalyco tap.
// Linux: npm install (sudo if system npm).
func resolveOpenCodeInstall(profile system.PlatformProfile) (CommandSequence, error) {
	switch profile.PackageManager {
	case "brew":
		return CommandSequence{
			{"brew", "install", "anomalyco/tap/opencode"},
		}, nil
	case "apt", "pacman", "dnf":
		if profile.NpmWritable {
			return CommandSequence{{"npm", "install", "-g", "opencode-ai"}}, nil
		}
		return CommandSequence{{"sudo", "npm", "install", "-g", "opencode-ai"}}, nil
	default:
		return nil, fmt.Errorf(
			"unsupported platform for opencode: os=%q distro=%q pm=%q",
			profile.OS, profile.LinuxDistro, profile.PackageManager,
		)
	}
}

// resolveEngramInstall returns the install commands for Engram.
// macOS: brew tap + brew install.
// Linux/other: returns error — callers should use direct binary download.
func resolveEngramInstall(profile system.PlatformProfile) (CommandSequence, error) {
	switch profile.PackageManager {
	case "brew":
		return CommandSequence{
			{"brew", "tap", "gentleman-programming/homebrew-tap"},
			{"brew", "install", "engram"},
		}, nil
	default:
		return nil, fmt.Errorf(
			"engram on %q/%q requires direct binary download",
			profile.OS, profile.PackageManager,
		)
	}
}
