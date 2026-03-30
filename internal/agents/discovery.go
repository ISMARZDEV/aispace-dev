package agents

import "os"

// InstalledAgent pairs an agent ID with its resolved config root directory.
type InstalledAgent struct {
	ID        string // model.AgentID value
	ConfigDir string // GlobalConfigDir value — guaranteed non-empty and exists on disk
}

// DiscoverInstalled returns agents whose GlobalConfigDir exists on disk.
// It performs a pure filesystem check — no subprocess spawning.
func DiscoverInstalled(reg *Registry, homeDir string) []InstalledAgent {
	var out []InstalledAgent

	for _, id := range reg.SupportedAgents() {
		adapter, ok := reg.Get(id)
		if !ok {
			continue
		}

		dir := adapter.GlobalConfigDir(homeDir)
		if dir == "" {
			continue
		}

		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue
		}

		out = append(out, InstalledAgent{ID: string(id), ConfigDir: dir})
	}

	return out
}

// ConfigRootsForBackup returns deduplicated config root directories for all
// agents whose GlobalConfigDir exists on disk. Used to enumerate files for backup.
func ConfigRootsForBackup(reg *Registry, homeDir string) []string {
	installed := DiscoverInstalled(reg, homeDir)

	seen := make(map[string]struct{}, len(installed))
	dirs := make([]string, 0, len(installed))

	for _, a := range installed {
		if _, ok := seen[a.ConfigDir]; ok {
			continue
		}
		seen[a.ConfigDir] = struct{}{}
		dirs = append(dirs, a.ConfigDir)
	}

	return dirs
}
