package claude

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ismartz/aispace-setup/internal/system"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name            string
		lookPathPath    string
		lookPathErr     error
		stat            statResult
		wantInstalled   bool
		wantBinaryPath  string
		wantConfigPath  string
		wantConfigFound bool
		wantErr         bool
	}{
		{
			name:            "binary and config directory found",
			lookPathPath:    "/usr/local/bin/claude",
			stat:            statResult{isDir: true},
			wantInstalled:   true,
			wantBinaryPath:  "/usr/local/bin/claude",
			wantConfigPath:  filepath.Join("/tmp/home", ".claude"),
			wantConfigFound: true,
		},
		{
			name:            "binary missing and config missing",
			lookPathErr:     errors.New("not found"),
			stat:            statResult{err: os.ErrNotExist},
			wantInstalled:   false,
			wantBinaryPath:  "",
			wantConfigPath:  filepath.Join("/tmp/home", ".claude"),
			wantConfigFound: false,
		},
		{
			name:    "stat error bubbles up",
			stat:    statResult{err: errors.New("permission denied")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Adapter{
				lookPath: func(string) (string, error) { return tt.lookPathPath, tt.lookPathErr },
				statPath: func(string) statResult { return tt.stat },
			}

			installed, binaryPath, configPath, configFound, err := a.Detect(context.Background(), "/tmp/home")
			if (err != nil) != tt.wantErr {
				t.Fatalf("Detect() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if installed != tt.wantInstalled {
				t.Errorf("installed = %v, want %v", installed, tt.wantInstalled)
			}
			if binaryPath != tt.wantBinaryPath {
				t.Errorf("binaryPath = %q, want %q", binaryPath, tt.wantBinaryPath)
			}
			if configPath != tt.wantConfigPath {
				t.Errorf("configPath = %q, want %q", configPath, tt.wantConfigPath)
			}
			if configFound != tt.wantConfigFound {
				t.Errorf("configFound = %v, want %v", configFound, tt.wantConfigFound)
			}
		})
	}
}

func TestInstallCommand(t *testing.T) {
	a := NewAdapter()

	tests := []struct {
		name    string
		profile system.PlatformProfile
		want    [][]string
	}{
		{
			name:    "darwin — npm without sudo",
			profile: system.PlatformProfile{OS: "darwin", PackageManager: "brew"},
			want:    [][]string{{"npm", "install", "-g", "@anthropic-ai/claude-code"}},
		},
		{
			name:    "linux system npm — requires sudo",
			profile: system.PlatformProfile{OS: "linux", PackageManager: "apt", NpmWritable: false},
			want:    [][]string{{"sudo", "npm", "install", "-g", "@anthropic-ai/claude-code"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := a.InstallCommand(tt.profile)
			if err != nil {
				t.Fatalf("InstallCommand() error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPaths(t *testing.T) {
	a := NewAdapter()
	home := "/home/user"

	if got := a.GlobalConfigDir(home); got != home+"/.claude" {
		t.Errorf("GlobalConfigDir = %q", got)
	}
	if got := a.SystemPromptFile(home); got != home+"/.claude/CLAUDE.md" {
		t.Errorf("SystemPromptFile = %q", got)
	}
	if got := a.SkillsDir(home); got != home+"/.claude/skills" {
		t.Errorf("SkillsDir = %q", got)
	}
	if got := a.SettingsPath(home); got != home+"/.claude/settings.json" {
		t.Errorf("SettingsPath = %q", got)
	}
	if got := a.MCPConfigPath(home, "context7"); got != home+"/.claude/mcp/context7.json" {
		t.Errorf("MCPConfigPath = %q", got)
	}
}

func TestStrategies(t *testing.T) {
	a := NewAdapter()

	if a.SystemPromptStrategy() != "markdown_sections" {
		t.Errorf("unexpected SystemPromptStrategy: %q", a.SystemPromptStrategy())
	}
	if a.MCPStrategy() != "separate_mcp_files" {
		t.Errorf("unexpected MCPStrategy: %q", a.MCPStrategy())
	}
	if a.SupportsSlashCommands() {
		t.Error("Claude Code should not support slash commands")
	}
	if !a.SupportsSkills() {
		t.Error("Claude Code should support skills")
	}
}
