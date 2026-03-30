package agents_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/ismartz/aispace-setup/internal/agents"
	"github.com/ismartz/aispace-setup/internal/model"
	"github.com/ismartz/aispace-setup/internal/system"
)

// mockAdapter satisfies the Adapter interface for testing.
type mockAdapter struct {
	agent model.AgentID
}

func (m mockAdapter) Agent() model.AgentID      { return m.agent }
func (m mockAdapter) Tier() model.SupportTier   { return model.TierFull }
func (m mockAdapter) SupportsAutoInstall() bool { return true }
func (m mockAdapter) Detect(_ context.Context, _ string) (bool, string, string, bool, error) {
	return false, "", "", false, nil
}
func (m mockAdapter) InstallCommand(system.PlatformProfile) ([][]string, error) { return nil, nil }
func (m mockAdapter) GlobalConfigDir(_ string) string                           { return "" }
func (m mockAdapter) SystemPromptDir(_ string) string                           { return "" }
func (m mockAdapter) SystemPromptFile(_ string) string                          { return "" }
func (m mockAdapter) SkillsDir(_ string) string                                 { return "" }
func (m mockAdapter) SettingsPath(_ string) string                              { return "" }
func (m mockAdapter) SystemPromptStrategy() model.SystemPromptStrategy {
	return model.StrategyMarkdownSections
}
func (m mockAdapter) MCPStrategy() model.MCPStrategy          { return model.StrategySeparateMCPFiles }
func (m mockAdapter) MCPConfigPath(_ string, _ string) string { return "" }
func (m mockAdapter) SupportsOutputStyles() bool              { return false }
func (m mockAdapter) OutputStyleDir(_ string) string          { return "" }
func (m mockAdapter) SupportsSlashCommands() bool             { return false }
func (m mockAdapter) CommandsDir(_ string) string             { return "" }
func (m mockAdapter) SupportsSkills() bool                    { return true }
func (m mockAdapter) SupportsSystemPrompt() bool              { return true }
func (m mockAdapter) SupportsMCP() bool                       { return true }

func TestRegistrySupportedAgentsSorted(t *testing.T) {
	r, err := agents.NewRegistry(
		mockAdapter{agent: model.AgentOpenCode},
		mockAdapter{agent: model.AgentClaudeCode},
	)
	if err != nil {
		t.Fatalf("NewRegistry() error: %v", err)
	}

	want := []model.AgentID{model.AgentClaudeCode, model.AgentOpenCode}
	if !reflect.DeepEqual(r.SupportedAgents(), want) {
		t.Errorf("SupportedAgents() = %v, want %v", r.SupportedAgents(), want)
	}
}

func TestRegistryRejectsDuplicateAgent(t *testing.T) {
	_, err := agents.NewRegistry(
		mockAdapter{agent: model.AgentClaudeCode},
		mockAdapter{agent: model.AgentClaudeCode},
	)
	if err == nil {
		t.Fatal("expected duplicate adapter error")
	}
	if !errors.Is(err, agents.ErrDuplicateAdapter) {
		t.Errorf("error = %v, want ErrDuplicateAdapter", err)
	}
}

func TestFactoryReturnsMVPAdapters(t *testing.T) {
	registry, err := agents.NewMVPRegistry()
	if err != nil {
		t.Fatalf("NewMVPRegistry() error: %v", err)
	}

	if _, ok := registry.Get(model.AgentClaudeCode); !ok {
		t.Error("registry missing claude adapter")
	}
	if _, ok := registry.Get(model.AgentOpenCode); !ok {
		t.Error("registry missing opencode adapter")
	}
}

func TestFactoryRejectsUnsupportedAgent(t *testing.T) {
	_, err := agents.NewAdapter(model.AgentID("unknown-xyz"))
	if err == nil {
		t.Fatal("expected unsupported agent error")
	}
	if !errors.Is(err, agents.ErrAgentNotSupported) {
		t.Errorf("error = %v, want ErrAgentNotSupported", err)
	}
}
