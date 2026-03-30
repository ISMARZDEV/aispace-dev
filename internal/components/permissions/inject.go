package permissions

import (
	"fmt"
	"os"

	"github.com/ismartz/aispace-setup/internal/agents"
	"github.com/ismartz/aispace-setup/internal/filemerge"
	"github.com/ismartz/aispace-setup/internal/model"
)

type InjectionResult struct {
	Changed bool
	Files   []string
}

// claudePermissionsOverlay — bypassPermissions mode with deny rules for destructive ops and .env files
var claudePermissionsOverlay = []byte(`{
  "permissions": {
    "defaultMode": "bypassPermissions",
    "deny": [
      "Bash(rm -rf /)",
      "Bash(sudo rm -rf /)",
      "Bash(rm -rf ~)",
      "Bash(sudo rm -rf ~)",
      "Read(.env)",
      "Read(.env.*)",
      "Edit(.env)",
      "Edit(.env.*)"
    ]
  }
}
`)

// opencodePermissionsOverlay — granular bash/read allow/deny rules
var opencodePermissionsOverlay = []byte(`{
  "permission": {
    "bash": {
      "*": "allow",
      "git push --force *": "ask",
      "git reset --hard *": "ask",
      "rm -rf *": "ask"
    },
    "read": {
      "*": "allow",
      "*.env": "deny",
      "**/.env*": "deny",
      "**/secrets/**": "deny"
    }
  }
}
`)

// Inject writes permission guardrails into the agent's settings file.
// Claude Code: merges into ~/.claude/settings.json
// OpenCode: merges into ~/.config/opencode/opencode.json
func Inject(homeDir string, adapter agents.Adapter) (InjectionResult, error) {
	settingsPath := adapter.SettingsPath(homeDir)
	if settingsPath == "" {
		return InjectionResult{}, nil
	}

	overlay, err := overlayForAgent(adapter.Agent())
	if err != nil {
		return InjectionResult{}, err
	}

	existing, readErr := readJSONOrEmpty(settingsPath)
	if readErr != nil {
		return InjectionResult{}, fmt.Errorf("permissions: read settings %q: %w", settingsPath, readErr)
	}

	merged, mergeErr := filemerge.MergeJSONObjects(existing, overlay)
	if mergeErr != nil {
		return InjectionResult{}, fmt.Errorf("permissions: merge JSON for %q: %w", settingsPath, mergeErr)
	}

	result, writeErr := filemerge.WriteFileAtomic(settingsPath, merged, 0o644)
	if writeErr != nil {
		return InjectionResult{}, fmt.Errorf("permissions: write settings %q: %w", settingsPath, writeErr)
	}

	return InjectionResult{Changed: result.Changed, Files: []string{settingsPath}}, nil
}

func overlayForAgent(agent model.AgentID) ([]byte, error) {
	switch agent {
	case model.AgentClaudeCode:
		return claudePermissionsOverlay, nil
	case model.AgentOpenCode:
		return opencodePermissionsOverlay, nil
	default:
		return nil, fmt.Errorf("permissions: no overlay defined for agent %q", agent)
	}
}

func readJSONOrEmpty(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []byte("{}"), nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return []byte("{}"), nil
	}
	return data, nil
}
