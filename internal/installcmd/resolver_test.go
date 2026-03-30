package installcmd_test

import (
	"reflect"
	"testing"

	"github.com/ismartz/aispace-setup/internal/installcmd"
	"github.com/ismartz/aispace-setup/internal/model"
	"github.com/ismartz/aispace-setup/internal/system"
)

func TestResolveClaudeCodeInstall(t *testing.T) {
	r := installcmd.NewResolver()

	tests := []struct {
		name    string
		profile system.PlatformProfile
		want    [][]string
	}{
		{
			name:    "darwin uses npm without sudo",
			profile: system.PlatformProfile{OS: "darwin", PackageManager: "brew"},
			want:    [][]string{{"npm", "install", "-g", "@anthropic-ai/claude-code"}},
		},
		{
			name:    "linux system npm requires sudo",
			profile: system.PlatformProfile{OS: "linux", PackageManager: "apt", NpmWritable: false},
			want:    [][]string{{"sudo", "npm", "install", "-g", "@anthropic-ai/claude-code"}},
		},
		{
			name:    "linux with nvm skips sudo",
			profile: system.PlatformProfile{OS: "linux", PackageManager: "apt", NpmWritable: true},
			want:    [][]string{{"npm", "install", "-g", "@anthropic-ai/claude-code"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.ResolveAgentInstall(tt.profile, model.AgentClaudeCode)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveOpenCodeInstall(t *testing.T) {
	r := installcmd.NewResolver()

	tests := []struct {
		name    string
		profile system.PlatformProfile
		want    [][]string
		wantErr bool
	}{
		{
			name:    "darwin uses brew tap",
			profile: system.PlatformProfile{OS: "darwin", PackageManager: "brew"},
			want:    [][]string{{"brew", "install", "anomalyco/tap/opencode"}},
		},
		{
			name:    "ubuntu system npm requires sudo",
			profile: system.PlatformProfile{OS: "linux", LinuxDistro: system.LinuxDistroUbuntu, PackageManager: "apt", NpmWritable: false},
			want:    [][]string{{"sudo", "npm", "install", "-g", "opencode-ai"}},
		},
		{
			name:    "arch with nvm skips sudo",
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
			got, err := r.ResolveAgentInstall(tt.profile, model.AgentOpenCode)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
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

func TestResolveEngramInstall(t *testing.T) {
	r := installcmd.NewResolver()

	got, err := r.ResolveComponentInstall(system.PlatformProfile{OS: "darwin", PackageManager: "brew"}, model.ComponentEngram)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := [][]string{
		{"brew", "tap", "gentleman-programming/homebrew-tap"},
		{"brew", "install", "engram"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	// Linux without brew returns error.
	_, err = r.ResolveComponentInstall(system.PlatformProfile{OS: "linux", PackageManager: "apt"}, model.ComponentEngram)
	if err == nil {
		t.Error("expected error for linux engram install")
	}
}
