package claude

import "path/filepath"

// ConfigPath returns the Claude Code global config directory (~/.claude).
func ConfigPath(homeDir string) string {
	return filepath.Join(homeDir, ".claude")
}
