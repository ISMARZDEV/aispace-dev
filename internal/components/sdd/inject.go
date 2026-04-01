package sdd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ismartz/aispace-setup/assets"
	"github.com/ismartz/aispace-setup/internal/agents"
	"github.com/ismartz/aispace-setup/internal/filemerge"
	"github.com/ismartz/aispace-setup/internal/model"
)

type InjectionResult struct {
	Changed bool
	Files   []string
}

// sddSkillIDs lists all 9 SDD phase skill IDs in workflow order.
var sddSkillIDs = []model.SkillID{
	"sdd-init",
	"sdd-explore",
	"sdd-propose",
	"sdd-spec",
	"sdd-design",
	"sdd-tasks",
	"sdd-apply",
	"sdd-verify",
	"sdd-archive",
}

// Inject injects the SDD orchestrator prompt and all 9 SDD phase skill files.
// Claude Code (MarkdownSections): injects orchestrator as a named section in CLAUDE.md.
// OpenCode (FileReplace): writes the full AGENTS.md with orchestrator content.
// If assigns is non-nil, a model assignment table is appended to the orchestrator section.
func Inject(homeDir string, adapter agents.Adapter, assigns model.ClaudeModelAssignments) (InjectionResult, error) {
	files := make([]string, 0, 12)
	changed := false

	// 1. Inject orchestrator into system prompt.
	if adapter.SupportsSystemPrompt() {
		orchFiles, orchChanged, err := injectOrchestrator(homeDir, adapter, assigns)
		if err != nil {
			return InjectionResult{}, err
		}
		files = append(files, orchFiles...)
		changed = changed || orchChanged
	}

	// 2. Write SDD skill files.
	if adapter.SupportsSkills() {
		skillFiles, skillChanged, err := injectSDDSkills(homeDir, adapter)
		if err != nil {
			return InjectionResult{}, err
		}
		files = append(files, skillFiles...)
		changed = changed || skillChanged
	}

	return InjectionResult{Changed: changed, Files: files}, nil
}

func injectOrchestrator(homeDir string, adapter agents.Adapter, assigns model.ClaudeModelAssignments) ([]string, bool, error) {
	assetPath := orchestratorAssetPath(adapter.Agent())
	raw, err := assets.Read(assetPath)
	if err != nil {
		return nil, false, fmt.Errorf("sdd: read orchestrator asset %q: %w", assetPath, err)
	}
	// Strip outer section markers that the asset includes for standalone readability.
	content := stripSectionWrapper(raw, "sdd-orchestrator")

	// Append model assignment table for Claude Code agents.
	if len(assigns) > 0 && adapter.Agent() == model.AgentClaudeCode {
		content = content + "\n\n" + buildModelAssignmentTable(assigns)
	}

	promptPath := adapter.SystemPromptFile(homeDir)

	switch adapter.SystemPromptStrategy() {
	case model.StrategyMarkdownSections:
		existing, err := readFileOrEmpty(promptPath)
		if err != nil {
			return nil, false, fmt.Errorf("sdd: read %q: %w", promptPath, err)
		}
		updated := filemerge.InjectMarkdownSection(existing, "sdd-orchestrator", content)
		result, err := filemerge.WriteFileAtomic(promptPath, []byte(updated), 0o644)
		if err != nil {
			return nil, false, fmt.Errorf("sdd: write %q: %w", promptPath, err)
		}
		return []string{promptPath}, result.Changed, nil

	case model.StrategyFileReplace:
		existing, err := readFileOrEmpty(promptPath)
		if err != nil {
			return nil, false, fmt.Errorf("sdd: read %q: %w", promptPath, err)
		}
		updated := filemerge.InjectMarkdownSection(existing, "sdd-orchestrator", content) //nolint:govet
		result, err := filemerge.WriteFileAtomic(promptPath, []byte(updated), 0o644)
		if err != nil {
			return nil, false, fmt.Errorf("sdd: write %q: %w", promptPath, err)
		}
		return []string{promptPath}, result.Changed, nil

	default:
		return nil, false, fmt.Errorf("sdd: unsupported strategy %q for agent %q", adapter.SystemPromptStrategy(), adapter.Agent())
	}
}

func injectSDDSkills(homeDir string, adapter agents.Adapter) ([]string, bool, error) {
	skillDir := adapter.SkillsDir(homeDir)
	if skillDir == "" {
		return nil, false, nil
	}

	files := make([]string, 0, len(sddSkillIDs))
	changed := false

	for _, id := range sddSkillIDs {
		assetPath := "skills/" + string(id) + "/SKILL.md"
		content, err := assets.Read(assetPath)
		if err != nil {
			// SDD skill asset missing — skip with warning, don't abort.
			continue
		}

		path := filepath.Join(skillDir, string(id), "SKILL.md")
		result, err := filemerge.WriteFileAtomic(path, []byte(content), 0o644)
		if err != nil {
			return nil, false, fmt.Errorf("sdd skill %q: write failed: %w", id, err)
		}
		changed = changed || result.Changed
		files = append(files, path)
	}

	return files, changed, nil
}

// orchestratorAssetPath returns the embedded asset path for the SDD orchestrator
// based on the agent type.
func orchestratorAssetPath(agent model.AgentID) string {
	switch agent {
	case model.AgentOpenCode:
		return "opencode/sdd-orchestrator.md"
	default:
		return "claude/sdd-orchestrator.md"
	}
}

// stripSectionWrapper removes the outer <!-- ai-setup:id --> / <!-- /ai-setup:id --> markers
// from asset files that include them for standalone readability.
func stripSectionWrapper(content, sectionID string) string {
	open := fmt.Sprintf("<!-- ai-setup:%s -->", sectionID)
	close := fmt.Sprintf("<!-- /ai-setup:%s -->", sectionID)

	start := strings.Index(content, open)
	end := strings.Index(content, close)
	if start == -1 || end == -1 || end <= start {
		return content
	}

	return strings.TrimSpace(content[start+len(open) : end])
}

// buildModelAssignmentTable renders a markdown table of SDD phase → model alias.
func buildModelAssignmentTable(assigns model.ClaudeModelAssignments) string {
	var b strings.Builder
	b.WriteString("### Model Assignments\n\n")
	b.WriteString("| Phase | Model |\n")
	b.WriteString("|-------|-------|\n")
	for _, phase := range model.AllSDDPhases() {
		alias := assigns[phase]
		if alias == "" {
			alias = model.ClaudeModelSonnet
		}
		b.WriteString(fmt.Sprintf("| %s | %s |\n", phase, alias))
	}
	return b.String()
}

func readFileOrEmpty(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}
