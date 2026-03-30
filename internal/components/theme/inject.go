package theme

import (
	"fmt"
	"os"

	"github.com/ismartz/aispace-setup/internal/agents"
	"github.com/ismartz/aispace-setup/internal/filemerge"
)

type InjectionResult struct {
	Changed bool
	Files   []string
}

// ayuDarkOverlay — Ayu Dark theme setting for Claude Code
var ayuDarkOverlay = []byte(`{
  "theme": "ayu-dark"
}
`)

// Inject merges the Ayu Dark theme setting into the agent's settings file.
// Only agents that have a non-empty SettingsPath will be modified.
func Inject(homeDir string, adapter agents.Adapter) (InjectionResult, error) {
	settingsPath := adapter.SettingsPath(homeDir)
	if settingsPath == "" {
		return InjectionResult{}, nil
	}

	existing, err := readJSONOrEmpty(settingsPath)
	if err != nil {
		return InjectionResult{}, fmt.Errorf("theme: read settings %q: %w", settingsPath, err)
	}

	merged, err := filemerge.MergeJSONObjects(existing, ayuDarkOverlay)
	if err != nil {
		return InjectionResult{}, fmt.Errorf("theme: merge JSON for %q: %w", settingsPath, err)
	}

	result, err := filemerge.WriteFileAtomic(settingsPath, merged, 0o644)
	if err != nil {
		return InjectionResult{}, fmt.Errorf("theme: write settings %q: %w", settingsPath, err)
	}

	return InjectionResult{Changed: result.Changed, Files: []string{settingsPath}}, nil
}

var readJSONOrEmpty = func(path string) ([]byte, error) {
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
