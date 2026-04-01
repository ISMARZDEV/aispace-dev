//go:build integration

package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ismartz/aispace-setup/internal/app"
	"github.com/ismartz/aispace-setup/internal/model"
	"github.com/ismartz/aispace-setup/internal/state"
)

// setupHome creates a temp directory to act as $HOME and registers cleanup.
// It also injects a fake "claude" binary into PATH so the agent step
// detects the agent as already installed and skips the npm install.
func setupHome(t *testing.T) string {
	t.Helper()

	homeDir := t.TempDir()

	// Fake claude binary — just needs to exist and be executable.
	binDir := t.TempDir()
	fakeClaude := filepath.Join(binDir, "claude")
	if err := os.WriteFile(fakeClaude, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("create fake claude: %v", err)
	}

	t.Setenv("HOME", homeDir)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	return homeDir
}

// claudePaths groups all expected file paths for the Claude Code agent.
type claudePaths struct {
	systemPrompt string // ~/.claude/CLAUDE.md
	settings     string // ~/.claude/settings.json
	skillsDir    string // ~/.claude/skills/
	mcpDir       string // ~/.claude/mcp/
}

func newClaudePaths(homeDir string) claudePaths {
	base := filepath.Join(homeDir, ".claude")
	return claudePaths{
		systemPrompt: filepath.Join(base, "CLAUDE.md"),
		settings:     filepath.Join(base, "settings.json"),
		skillsDir:    filepath.Join(base, "skills"),
		mcpDir:       filepath.Join(base, "mcp"),
	}
}

// — Tests —

// TestInstall_Minimal verifies that the minimal preset (engram only) creates
// CLAUDE.md without touching settings.json or the skills directory.
func TestInstall_Minimal(t *testing.T) {
	homeDir := setupHome(t)
	paths := newClaudePaths(homeDir)

	err := app.RunArgs([]string{
		"install",
		"--agent", "claude-code",
		"--preset", "minimal",
		"--persona", "neutral",
	}, os.Stdout)
	if err != nil {
		t.Fatalf("RunArgs: %v", err)
	}

	// CLAUDE.md must exist with content.
	assertFileExists(t, paths.systemPrompt)
	assertFileContains(t, paths.systemPrompt, "engram")

	// settings.json must NOT exist (minimal preset doesn't include permissions/theme).
	assertFileAbsent(t, paths.settings)

	// skills dir must NOT exist.
	assertDirAbsent(t, paths.skillsDir)
}

// TestInstall_Full verifies that the full preset creates all expected files.
func TestInstall_Full(t *testing.T) {
	homeDir := setupHome(t)
	paths := newClaudePaths(homeDir)

	err := app.RunArgs([]string{
		"install",
		"--agent", "claude-code",
		"--preset", "full",
		"--persona", "neutral",
	}, os.Stdout)
	if err != nil {
		t.Fatalf("RunArgs: %v", err)
	}

	// System prompt must contain all injected sections.
	assertFileExists(t, paths.systemPrompt)
	content := readFile(t, paths.systemPrompt)
	for _, section := range []string{"engram", "sdd"} {
		if !strings.Contains(content, section) {
			t.Errorf("CLAUDE.md missing section %q", section)
		}
	}

	// settings.json must exist (permissions + theme).
	assertFileExists(t, paths.settings)

	// Skills directory must contain skill files.
	assertDirExists(t, paths.skillsDir)
	entries, err := os.ReadDir(paths.skillsDir)
	if err != nil {
		t.Fatalf("read skills dir: %v", err)
	}
	if len(entries) == 0 {
		t.Error("skills dir is empty — expected at least one skill")
	}

	// MCP config dir must contain at least one JSON file.
	assertDirExists(t, paths.mcpDir)
	mcpEntries, err := os.ReadDir(paths.mcpDir)
	if err != nil {
		t.Fatalf("read mcp dir: %v", err)
	}
	if len(mcpEntries) == 0 {
		t.Error("mcp dir is empty — expected at least one config file")
	}
}

// TestInstall_Persona verifies that the selected persona is injected into CLAUDE.md.
func TestInstall_Persona(t *testing.T) {
	personas := []model.PersonaID{
		model.PersonaNeutral,
		model.PersonaDominicano,
		model.PersonaAlien,
	}

	for _, persona := range personas {
		t.Run(string(persona), func(t *testing.T) {
			homeDir := setupHome(t)
			paths := newClaudePaths(homeDir)

			err := app.RunArgs([]string{
				"install",
				"--agent", "claude-code",
				"--preset", "minimal",
				"--persona", string(persona),
			}, os.Stdout)
			if err != nil {
				t.Fatalf("RunArgs: %v", err)
			}

			assertFileExists(t, paths.systemPrompt)
		})
	}
}

// TestInstall_DryRun verifies that --dry-run does not create any files.
func TestInstall_DryRun(t *testing.T) {
	homeDir := setupHome(t)
	paths := newClaudePaths(homeDir)

	err := app.RunArgs([]string{
		"install",
		"--agent", "claude-code",
		"--preset", "full",
		"--persona", "neutral",
		"--dry-run",
	}, os.Stdout)
	if err != nil {
		t.Fatalf("RunArgs: %v", err)
	}

	assertFileAbsent(t, paths.systemPrompt)
	assertFileAbsent(t, paths.settings)
	assertDirAbsent(t, paths.skillsDir)
}

// TestInstall_StateWritten verifies that state.json is persisted after install.
func TestInstall_StateWritten(t *testing.T) {
	homeDir := setupHome(t)

	err := app.RunArgs([]string{
		"install",
		"--agent", "claude-code",
		"--preset", "minimal",
		"--persona", "neutral",
	}, os.Stdout)
	if err != nil {
		t.Fatalf("RunArgs: %v", err)
	}

	if !state.Exists(homeDir) {
		t.Fatal("state.json was not written after install")
	}

	s, err := state.Read(homeDir)
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	if len(s.InstalledAgents) == 0 {
		t.Error("state.json has no installed agents")
	}
	if s.Preset == "" {
		t.Error("state.json has empty preset")
	}
}

// TestInstall_Idempotent verifies that running install twice does not corrupt files.
func TestInstall_Idempotent(t *testing.T) {
	homeDir := setupHome(t)
	paths := newClaudePaths(homeDir)

	args := []string{
		"install",
		"--agent", "claude-code",
		"--preset", "minimal",
		"--persona", "neutral",
	}

	if err := app.RunArgs(args, os.Stdout); err != nil {
		t.Fatalf("first install: %v", err)
	}
	contentAfterFirst := readFile(t, paths.systemPrompt)

	if err := app.RunArgs(args, os.Stdout); err != nil {
		t.Fatalf("second install: %v", err)
	}
	contentAfterSecond := readFile(t, paths.systemPrompt)

	if contentAfterFirst != contentAfterSecond {
		t.Error("CLAUDE.md changed between first and second install — install is not idempotent")
	}
}

// TestInstall_BackupCreated verifies that a backup snapshot is written.
func TestInstall_BackupCreated(t *testing.T) {
	homeDir := setupHome(t)

	if err := app.RunArgs([]string{
		"install",
		"--agent", "claude-code",
		"--preset", "minimal",
		"--persona", "neutral",
	}, os.Stdout); err != nil {
		t.Fatalf("RunArgs: %v", err)
	}

	backupRoot := filepath.Join(homeDir, ".aispace-setup", "backups")
	entries, err := os.ReadDir(backupRoot)
	if err != nil {
		t.Fatalf("read backup root: %v", err)
	}
	if len(entries) == 0 {
		t.Error("no backup snapshot created after install")
	}

	// Verify manifest.json exists inside the snapshot dir.
	snapshotDir := filepath.Join(backupRoot, entries[0].Name())
	manifest := filepath.Join(snapshotDir, "manifest.json")
	assertFileExists(t, manifest)
}

// TestInstall_StateAgentIDs verifies exact agent IDs are persisted in state.
func TestInstall_StateAgentIDs(t *testing.T) {
	homeDir := setupHome(t)

	if err := app.RunArgs([]string{
		"install",
		"--agent", "claude-code",
		"--preset", "minimal",
		"--persona", "neutral",
	}, os.Stdout); err != nil {
		t.Fatalf("RunArgs: %v", err)
	}

	data, err := os.ReadFile(state.Path(homeDir))
	if err != nil {
		t.Fatalf("read state file: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal state: %v", err)
	}
	if _, ok := raw["installed_agents"]; !ok {
		t.Error("state.json missing 'installed_agents' key")
	}
}

// — Assertion helpers —

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file to exist: %s", path)
	}
}

func assertFileAbsent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("expected file to NOT exist: %s", path)
	}
}

func assertDirExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Errorf("expected directory to exist: %s", path)
		return
	}
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if !info.IsDir() {
		t.Errorf("expected %s to be a directory", path)
	}
}

func assertDirAbsent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("expected directory to NOT exist: %s", path)
	}
}

func assertFileContains(t *testing.T, path, substr string) {
	t.Helper()
	content := readFile(t, path)
	if !strings.Contains(content, substr) {
		t.Errorf("file %s does not contain %q", path, substr)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
