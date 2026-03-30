package persona

import (
	"fmt"
	"os"

	"github.com/ismartz/aispace-setup/assets"
	"github.com/ismartz/aispace-setup/internal/agents"
	"github.com/ismartz/aispace-setup/internal/filemerge"
	"github.com/ismartz/aispace-setup/internal/model"
)

type InjectionResult struct {
	Changed bool
	Files   []string
}

// Inject writes the persona content into the agent's system prompt file.
// For Claude Code (MarkdownSections): injects a named section into CLAUDE.md.
// For OpenCode (FileReplace): writes AGENTS.md entirely with persona as the base.
// If persona is PersonaCustom, nothing is done.
func Inject(homeDir string, adapter agents.Adapter, persona model.PersonaID) (InjectionResult, error) {
	if !adapter.SupportsSystemPrompt() {
		return InjectionResult{}, nil
	}
	if persona == model.PersonaCustom {
		return InjectionResult{}, nil
	}

	content, err := personaContent(persona)
	if err != nil {
		return InjectionResult{}, err
	}

	promptPath := adapter.SystemPromptFile(homeDir)

	switch adapter.SystemPromptStrategy() {
	case model.StrategyMarkdownSections:
		existing, err := readFileOrEmpty(promptPath)
		if err != nil {
			return InjectionResult{}, fmt.Errorf("read system prompt %q: %w", promptPath, err)
		}
		updated := filemerge.InjectMarkdownSection(existing, "persona", content)
		result, err := filemerge.WriteFileAtomic(promptPath, []byte(updated), 0o644)
		if err != nil {
			return InjectionResult{}, fmt.Errorf("write system prompt %q: %w", promptPath, err)
		}
		return InjectionResult{Changed: result.Changed, Files: []string{promptPath}}, nil

	case model.StrategyFileReplace:
		result, err := filemerge.WriteFileAtomic(promptPath, []byte(content), 0o644)
		if err != nil {
			return InjectionResult{}, fmt.Errorf("write system prompt %q: %w", promptPath, err)
		}
		return InjectionResult{Changed: result.Changed, Files: []string{promptPath}}, nil

	default:
		return InjectionResult{}, fmt.Errorf("persona injector: unsupported strategy %q for agent %q", adapter.SystemPromptStrategy(), adapter.Agent())
	}
}

// personaContent reads the persona markdown from embedded assets.
func personaContent(persona model.PersonaID) (string, error) {
	path := "personas/" + string(persona) + ".md"
	content, err := assets.Read(path)
	if err != nil {
		return "", fmt.Errorf("persona asset %q not found: %w", path, err)
	}
	return content, nil
}

// readFileOrEmpty reads path, returning empty string if the file does not exist.
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
