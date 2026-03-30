package opencode

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
			lookPathPath:    "/opt/homebrew/bin/opencode",
			stat:            statResult{isDir: true},
			wantInstalled:   true,
			wantBinaryPath:  "/opt/homebrew/bin/opencode",
			wantConfigPath:  filepath.Join("/tmp/home", ".config", "opencode"),
			wantConfigFound: true,
		},
		{
			name:            "binary missing and config missing",
			lookPathErr:     errors.New("not found"),
			stat:            statResult{err: os.ErrNotExist},
			wantInstalled:   false,
			wantBinaryPath:  "",
			wantConfigPath:  filepath.Join("/tmp/home", ".config", "opencode"),
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
		wantErr bool
	}{
		{
			name:    "darwin — brew tap",
			profile: system.PlatformProfile{OS: "darwin", PackageManager: "brew"},
			want:    [][]string{{"brew", "install", "anomalyco/tap/opencode"}},
		},
		{
			name:    "ubuntu system npm — requires sudo",
			profile: system.PlatformProfile{OS: "linux", LinuxDistro: system.LinuxDistroUbuntu, PackageManager: "apt", NpmWritable: false},
			want:    [][]string{{"sudo", "npm", "install", "-g", "opencode-ai"}},
		},
		{
			name:    "arch with nvm — skips sudo",
			profile: system.PlatformProfile{OS: "linux", LinuxDistro: system.LinuxDistroArch, PackageManager: "pacman", NpmWritable: true},
			want:    [][]string{{"npm", "install", "-g", "opencode-ai"}},
		},
		{
			name:    "unsupported package manager returns error",
			profile: system.PlatformProfile{OS: "linux", PackageManager: "zypper"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := a.InstallCommand(tt.profile)
			if (err != nil) != tt.wantErr {
				t.Fatalf("InstallCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
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

	if got := a.GlobalConfigDir(home); got != home+"/.config/opencode" {
		t.Errorf("GlobalConfigDir = %q", got)
	}
	if got := a.SystemPromptFile(home); got != home+"/.config/opencode/AGENTS.md" {
		t.Errorf("SystemPromptFile = %q", got)
	}
	if got := a.SkillsDir(home); got != home+"/.config/opencode/skills" {
		t.Errorf("SkillsDir = %q", got)
	}
	if got := a.SettingsPath(home); got != home+"/.config/opencode/opencode.json" {
		t.Errorf("SettingsPath = %q", got)
	}
	if got := a.CommandsDir(home); got != home+"/.config/opencode/commands" {
		t.Errorf("CommandsDir = %q", got)
	}
}

func TestStrategies(t *testing.T) {
	a := NewAdapter()

	if a.SystemPromptStrategy() != "file_replace" {
		t.Errorf("unexpected SystemPromptStrategy: %q", a.SystemPromptStrategy())
	}
	if a.MCPStrategy() != "merge_into_settings" {
		t.Errorf("unexpected MCPStrategy: %q", a.MCPStrategy())
	}
	if !a.SupportsSlashCommands() {
		t.Error("OpenCode should support slash commands")
	}
	if a.SupportsOutputStyles() {
		t.Error("OpenCode should not support output styles")
	}
}
