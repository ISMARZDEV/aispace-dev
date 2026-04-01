package engram

import (
	"fmt"
	"os"
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

// Inject writes the Engram memory protocol into the agent's system prompt.
// For Claude Code: injects a named section into CLAUDE.md.
// For OpenCode: appends to AGENTS.md (FileReplace agents manage the full file).
func Inject(homeDir string, adapter agents.Adapter) (InjectionResult, error) {
	if !adapter.SupportsSystemPrompt() {
		return InjectionResult{}, nil
	}

	raw, err := assets.Read("claude/engram-protocol.md")
	if err != nil {
		return InjectionResult{}, fmt.Errorf("engram: read protocol asset: %w", err)
	}
	// The asset file wraps content in ai-setup section markers for standalone readability.
	// Strip them before passing to InjectMarkdownSection, which adds its own markers.
	protocol := stripSectionWrapper(raw, "engram-protocol")

	promptPath := adapter.SystemPromptFile(homeDir)

	switch adapter.SystemPromptStrategy() {
	case model.StrategyMarkdownSections:
		existing, err := readFileOrEmpty(promptPath)
		if err != nil {
			return InjectionResult{}, fmt.Errorf("engram: read %q: %w", promptPath, err)
		}
		updated := filemerge.InjectMarkdownSection(existing, "engram-protocol", protocol)
		result, err := filemerge.WriteFileAtomic(promptPath, []byte(updated), 0o644)
		if err != nil {
			return InjectionResult{}, fmt.Errorf("engram: write %q: %w", promptPath, err)
		}
		return InjectionResult{Changed: result.Changed, Files: []string{promptPath}}, nil

	case model.StrategyFileReplace:
		// For FileReplace agents, persona already wrote the file.
		// Read existing and append the protocol section.
		existing, err := readFileOrEmpty(promptPath)
		if err != nil {
			return InjectionResult{}, fmt.Errorf("engram: read %q: %w", promptPath, err)
		}
		updated := filemerge.InjectMarkdownSection(existing, "engram-protocol", protocol)
		result, err := filemerge.WriteFileAtomic(promptPath, []byte(updated), 0o644)
		if err != nil {
			return InjectionResult{}, fmt.Errorf("engram: write %q: %w", promptPath, err)
		}
		return InjectionResult{Changed: result.Changed, Files: []string{promptPath}}, nil

	default:
		return InjectionResult{}, fmt.Errorf("engram: unsupported strategy %q for agent %q", adapter.SystemPromptStrategy(), adapter.Agent())
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

	inner := content[start+len(open) : end]
	return strings.TrimSpace(inner)
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
