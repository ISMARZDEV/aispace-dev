package mcp

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

// Inject configures MCP servers for the given agent.
// The injection strategy depends on the agent's MCPStrategy:
//   - SeparateMCPFiles (Claude Code): writes one JSON file per server
//   - MergeIntoSettings (OpenCode): merges all MCP config into the main settings file
func Inject(homeDir string, adapter agents.Adapter) (InjectionResult, error) {
	if !adapter.SupportsMCP() {
		return InjectionResult{}, nil
	}

	switch adapter.MCPStrategy() {
	case model.StrategySeparateMCPFiles:
		return injectSeparateFiles(homeDir, adapter)
	case model.StrategyMergeIntoSettings:
		return injectMergeIntoSettings(homeDir, adapter)
	default:
		return InjectionResult{}, fmt.Errorf("mcp injector: unsupported strategy %q for agent %q", adapter.MCPStrategy(), adapter.Agent())
	}
}

// injectSeparateFiles writes one JSON file per MCP server (Claude Code strategy).
func injectSeparateFiles(homeDir string, adapter agents.Adapter) (InjectionResult, error) {
	servers := mcpServersForAgent(adapter.Agent())
	files := make([]string, 0, len(servers))
	changed := false

	for _, serverName := range servers {
		config := claudeSeparateMCPConfig(serverName)
		if config == nil {
			continue
		}

		path := adapter.MCPConfigPath(homeDir, serverName)
		result, err := filemerge.WriteFileAtomic(path, config, 0o644)
		if err != nil {
			return InjectionResult{}, fmt.Errorf("mcp: write config for %q: %w", serverName, err)
		}
		changed = changed || result.Changed
		files = append(files, path)
	}

	return InjectionResult{Changed: changed, Files: files}, nil
}

// injectMergeIntoSettings merges all MCP config into the agent's main settings file (OpenCode strategy).
func injectMergeIntoSettings(homeDir string, adapter agents.Adapter) (InjectionResult, error) {
	settingsPath := adapter.SettingsPath(homeDir)

	existing, err := readJSONOrEmpty(settingsPath)
	if err != nil {
		return InjectionResult{}, fmt.Errorf("mcp: read settings %q: %w", settingsPath, err)
	}

	overlay := opencodeMCPOverlay()
	merged, err := filemerge.MergeJSONObjects(existing, overlay)
	if err != nil {
		return InjectionResult{}, fmt.Errorf("mcp: merge JSON for %q: %w", settingsPath, err)
	}

	result, err := filemerge.WriteFileAtomic(settingsPath, merged, 0o644)
	if err != nil {
		return InjectionResult{}, fmt.Errorf("mcp: write settings %q: %w", settingsPath, err)
	}

	return InjectionResult{Changed: result.Changed, Files: []string{settingsPath}}, nil
}

// readJSONOrEmpty reads JSON from path, returning []byte("{}") if file does not exist.
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
