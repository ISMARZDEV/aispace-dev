package model

// AgentID identifies a supported AI agent.
type AgentID string

const (
	AgentClaudeCode AgentID = "claude-code"
	AgentOpenCode   AgentID = "opencode"
)

// ComponentID identifies an installable component.
type ComponentID string

const (
	ComponentEngram     ComponentID = "engram"
	ComponentSDD        ComponentID = "sdd"
	ComponentSkills     ComponentID = "skills"
	ComponentContext7   ComponentID = "context7"
	ComponentPersona    ComponentID = "persona"
	ComponentPermission ComponentID = "permissions"
	ComponentTheme      ComponentID = "theme"
	ComponentAISpace    ComponentID = "aispace"
)

// PersonaID identifies a persona.
type PersonaID string

const (
	PersonaNeutral    PersonaID = "neutral"
	PersonaDominicano PersonaID = "dominicano"
	PersonaAlien      PersonaID = "alien"
	PersonaCustom     PersonaID = "custom" // user keeps their own config, nothing injected
)

// SkillID identifies a skill file (maps to assets/skills/<id>/SKILL.md).
type SkillID string

const (
	SkillJudgmentDay   SkillID = "judgment-day"
	SkillBranchPR      SkillID = "branch-pr"
	SkillIssueCreation SkillID = "issue-creation"
)

// PresetID identifies a predefined component bundle.
type PresetID string

const (
	PresetFull    PresetID = "full"
	PresetCore    PresetID = "core"
	PresetMinimal PresetID = "minimal"
	PresetCustom  PresetID = "custom"
)

// SDDModeID identifies the SDD orchestration mode.
type SDDModeID string

const (
	SDDModeSingle SDDModeID = "single" // one agent handles all phases
	SDDModeMulti  SDDModeID = "multi"  // sub-agents per phase
)

// SupportTier classifies how well an agent is supported.
type SupportTier string

const (
	TierFull    SupportTier = "full"    // all features supported
	TierPartial SupportTier = "partial" // subset of features supported
)

// SystemPromptStrategy defines how an agent accepts system prompt injections.
type SystemPromptStrategy string

const (
	// StrategyMarkdownSections injects HTML-comment-delimited sections into CLAUDE.md.
	StrategyMarkdownSections SystemPromptStrategy = "markdown_sections"
	// StrategyFileReplace writes the system prompt file from scratch (opencode overlays).
	StrategyFileReplace SystemPromptStrategy = "file_replace"
)

// MCPStrategy defines how an agent stores MCP server config.
type MCPStrategy string

const (
	// StrategySeparateMCPFiles writes .mcp.json files (Claude Code).
	StrategySeparateMCPFiles MCPStrategy = "separate_mcp_files"
	// StrategyMergeIntoSettings merges MCP config into agent's main settings file.
	StrategyMergeIntoSettings MCPStrategy = "merge_into_settings"
)

// Selection holds the user's install choices before resolution.
type Selection struct {
	Agents     []AgentID
	Components []ComponentID
	Persona    PersonaID
	Preset     PresetID
	SDDMode    SDDModeID
	StrictTDD  bool
}

// InstallState is the persisted state written after a successful install.
type InstallState struct {
	InstalledAgents []AgentID     `json:"installed_agents"`
	Persona         PersonaID     `json:"persona"`
	Preset          PresetID      `json:"preset"`
	SDDMode         SDDModeID     `json:"sdd_mode"`
	StrictTDD       bool          `json:"strict_tdd"`
	Components      []ComponentID `json:"components"`
}
