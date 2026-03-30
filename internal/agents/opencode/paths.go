package opencode

import "path/filepath"

// ConfigPath returns the OpenCode global config directory (~/.config/opencode).
func ConfigPath(homeDir string) string {
	return filepath.Join(homeDir, ".config", "opencode")
}
