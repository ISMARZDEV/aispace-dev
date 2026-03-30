package opencode

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ismartz/aispace-setup/internal/installcmd"
	"github.com/ismartz/aispace-setup/internal/model"
	"github.com/ismartz/aispace-setup/internal/system"
)

// LookPathOverride allows tests to inject a custom LookPath function.
var LookPathOverride = exec.LookPath

type statResult struct {
	isDir bool
	err   error
}

// Adapter implements the agents.Adapter interface for OpenCode.
type Adapter struct {
	lookPath func(string) (string, error)
	statPath func(string) statResult
	resolver installcmd.Resolver
}

// NewAdapter creates a production Adapter with real OS calls.
func NewAdapter() *Adapter {
	return &Adapter{
		lookPath: LookPathOverride,
		statPath: defaultStat,
		resolver: installcmd.NewResolver(),
	}
}

// --- Identity ---

func (a *Adapter) Agent() model.AgentID {
	return model.AgentOpenCode
}

func (a *Adapter) Tier() model.SupportTier {
	return model.TierFull
}

// --- Detection ---

// Detect checks whether the opencode binary is on PATH and whether ~/.config/opencode exists.
func (a *Adapter) Detect(_ context.Context, homeDir string) (bool, string, string, bool, error) {
	configPath := ConfigPath(homeDir)

	binaryPath, err := a.lookPath("opencode")
	installed := err == nil

	stat := a.statPath(configPath)
	if stat.err != nil {
		if os.IsNotExist(stat.err) {
			return installed, binaryPath, configPath, false, nil
		}
		return false, "", "", false, stat.err
	}

	return installed, binaryPath, configPath, stat.isDir, nil
}

// --- Installation ---

func (a *Adapter) SupportsAutoInstall() bool {
	return true
}

func (a *Adapter) InstallCommand(profile system.PlatformProfile) ([][]string, error) {
	return a.resolver.ResolveAgentInstall(profile, a.Agent())
}

// --- Config paths ---

func (a *Adapter) GlobalConfigDir(homeDir string) string {
	return filepath.Join(homeDir, ".config", "opencode")
}

func (a *Adapter) SystemPromptDir(homeDir string) string {
	return filepath.Join(homeDir, ".config", "opencode")
}

// OpenCode uses AGENTS.md as its system prompt file.
func (a *Adapter) SystemPromptFile(homeDir string) string {
	return filepath.Join(homeDir, ".config", "opencode", "AGENTS.md")
}

func (a *Adapter) SkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".config", "opencode", "skills")
}

func (a *Adapter) SettingsPath(homeDir string) string {
	return filepath.Join(homeDir, ".config", "opencode", "opencode.json")
}

// --- Config strategies ---

// OpenCode uses FileReplace — writes the full system prompt file (AGENTS.md) from scratch.
// This is why Persona must run before SDD (soft ordering constraint).
func (a *Adapter) SystemPromptStrategy() model.SystemPromptStrategy {
	return model.StrategyFileReplace
}

// OpenCode merges MCP config into opencode.json rather than using separate files.
func (a *Adapter) MCPStrategy() model.MCPStrategy {
	return model.StrategyMergeIntoSettings
}

// --- MCP ---

// MCPConfigPath returns the main settings file for OpenCode (MCP is merged into it).
func (a *Adapter) MCPConfigPath(homeDir string, _ string) string {
	return filepath.Join(homeDir, ".config", "opencode", "opencode.json")
}

// --- Optional capabilities ---

func (a *Adapter) SupportsOutputStyles() bool {
	return false
}

func (a *Adapter) OutputStyleDir(_ string) string {
	return ""
}

// OpenCode supports slash commands via ~/.config/opencode/commands/.
func (a *Adapter) SupportsSlashCommands() bool {
	return true
}

func (a *Adapter) CommandsDir(homeDir string) string {
	return filepath.Join(homeDir, ".config", "opencode", "commands")
}

func (a *Adapter) SupportsSkills() bool {
	return true
}

func (a *Adapter) SupportsSystemPrompt() bool {
	return true
}

func (a *Adapter) SupportsMCP() bool {
	return true
}

func defaultStat(path string) statResult {
	info, err := os.Stat(path)
	if err != nil {
		return statResult{err: err}
	}
	return statResult{isDir: info.IsDir()}
}
