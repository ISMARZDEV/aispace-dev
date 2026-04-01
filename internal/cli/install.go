package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ismartz/aispace-setup/internal/agents"
	"github.com/ismartz/aispace-setup/internal/backup"
	"github.com/ismartz/aispace-setup/internal/components/engram"
	"github.com/ismartz/aispace-setup/internal/components/mcp"
	"github.com/ismartz/aispace-setup/internal/components/permissions"
	"github.com/ismartz/aispace-setup/internal/components/persona"
	"github.com/ismartz/aispace-setup/internal/components/sdd"
	"github.com/ismartz/aispace-setup/internal/components/skills"
	"github.com/ismartz/aispace-setup/internal/components/theme"
	"github.com/ismartz/aispace-setup/internal/model"
	"github.com/ismartz/aispace-setup/internal/pipeline"
	"github.com/ismartz/aispace-setup/internal/planner"
	"github.com/ismartz/aispace-setup/internal/state"
	"github.com/ismartz/aispace-setup/internal/system"
	"github.com/ismartz/aispace-setup/internal/verify"
)

// AppVersion is stamped into backup manifests. Set by app.go at startup.
var AppVersion = "dev"

var osUserHomeDir = os.UserHomeDir

// InstallFlags parsed from CLI args.
type InstallFlags struct {
	Agents     []string
	Components []string
	Persona    string
	Preset     string
	SDDMode    string
	StrictTDD  bool
	DryRun     bool
}

// InstallResult holds the outcome of an install execution.
type InstallResult struct {
	Selection model.Selection
	Resolved  planner.ResolvedPlan
	Plan      pipeline.StagePlan
	Execution pipeline.ExecutionResult
	Verify    []verify.CheckResult
	DryRun    bool
}

// ParseInstallFlags parses CLI args for the install subcommand.
func ParseInstallFlags(args []string) (InstallFlags, error) {
	var opts InstallFlags

	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.SetOutput(ioDiscard{})
	registerListFlag(fs, "agent", &opts.Agents)
	registerListFlag(fs, "agents", &opts.Agents)
	registerListFlag(fs, "component", &opts.Components)
	registerListFlag(fs, "components", &opts.Components)
	fs.StringVar(&opts.Persona, "persona", "", "persona to apply (neutral|dominicano|alien|custom)")
	fs.StringVar(&opts.Preset, "preset", "", "preset bundle (full|core|minimal|custom)")
	fs.StringVar(&opts.SDDMode, "sdd-mode", "", "SDD orchestrator mode: single or multi (default: single)")
	fs.BoolVar(&opts.StrictTDD, "strict-tdd", false, "enable strict TDD mode for SDD agents")
	fs.BoolVar(&opts.DryRun, "dry-run", false, "preview plan without executing")

	if err := fs.Parse(args); err != nil {
		return InstallFlags{}, err
	}

	if fs.NArg() > 0 {
		return InstallFlags{}, fmt.Errorf("unexpected install argument %q", fs.Arg(0))
	}

	return opts, nil
}

// BuildInstallSelection converts InstallFlags → model.Selection.
func BuildInstallSelection(flags InstallFlags, agentIDs []model.AgentID) (model.Selection, error) {
	sel := model.Selection{
		Agents:    agentIDs,
		Persona:   model.PersonaID(flags.Persona),
		Preset:    model.PresetID(flags.Preset),
		SDDMode:   model.SDDModeID(flags.SDDMode),
		StrictTDD: flags.StrictTDD,
	}

	if len(flags.Components) > 0 {
		components := make([]model.ComponentID, 0, len(flags.Components))
		for _, c := range flags.Components {
			components = append(components, model.ComponentID(c))
		}
		sel.Components = components
	}

	return NormalizeSelection(sel)
}

// RunInstall is the top-level install entry point.
func RunInstall(args []string, profile system.PlatformProfile) (InstallResult, error) {
	flags, err := ParseInstallFlags(args)
	if err != nil {
		return InstallResult{}, err
	}

	homeDir, err := osUserHomeDir()
	if err != nil {
		return InstallResult{}, fmt.Errorf("resolve user home directory: %w", err)
	}

	var agentIDs []model.AgentID
	if len(flags.Agents) > 0 {
		agentIDs = asAgentIDs(flags.Agents)
	} else {
		agentIDs = discoverAgents(homeDir)
	}
	agentIDs = unique(agentIDs)

	sel, err := BuildInstallSelection(flags, agentIDs)
	if err != nil {
		return InstallResult{}, err
	}

	resolved, err := planner.NewResolver(planner.DefaultGraph()).Resolve(sel)
	if err != nil {
		return InstallResult{}, fmt.Errorf("resolve plan: %w", err)
	}

	rt, err := newInstallRuntime(homeDir, sel, resolved)
	if err != nil {
		return InstallResult{}, err
	}

	stagePlan := rt.stagePlan()

	result := InstallResult{
		Selection: sel,
		Resolved:  resolved,
		Plan:      stagePlan,
		DryRun:    flags.DryRun,
	}

	if flags.DryRun {
		return result, nil
	}

	result.Execution = ExecuteInstallPlan(sel, resolved, profile, nil)
	if result.Execution.Err != nil {
		return result, fmt.Errorf("execute install pipeline: %w", result.Execution.Err)
	}

	result.Verify = runPostInstallVerification(homeDir, sel, resolved)
	if verify.AnyFailed(result.Verify) {
		return result, fmt.Errorf("post-install verification failed")
	}

	_ = state.Write(homeDir, model.InstallState{
		InstalledAgents: sel.Agents,
		Persona:         sel.Persona,
		Preset:          sel.Preset,
		SDDMode:         sel.SDDMode,
		StrictTDD:       sel.StrictTDD,
		Components:      sel.Components,
	})

	return result, nil
}

// ExecuteInstallPlan builds and runs the pipeline for an install.
// Called by RunInstall (CLI) and app.tuiExecute (TUI).
func ExecuteInstallPlan(
	selection model.Selection,
	resolved planner.ResolvedPlan,
	_ system.PlatformProfile,
	onProgress pipeline.ProgressFunc,
) pipeline.ExecutionResult {
	homeDir, err := osUserHomeDir()
	if err != nil {
		return pipeline.ExecutionResult{Err: fmt.Errorf("resolve home directory: %w", err)}
	}

	rt, err := newInstallRuntime(homeDir, selection, resolved)
	if err != nil {
		return pipeline.ExecutionResult{Err: err}
	}

	opts := []pipeline.OrchestratorOption{}
	if onProgress != nil {
		opts = append(opts, pipeline.WithProgressFunc(onProgress))
	}

	return pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy(), opts...).Execute(rt.stagePlan())
}

// runtimeState holds state shared across pipeline steps.
type runtimeState struct {
	manifest backup.Manifest
}

// installRuntime builds the StagePlan for an install.
type installRuntime struct {
	homeDir    string
	selection  model.Selection
	resolved   planner.ResolvedPlan
	backupRoot string
	state      *runtimeState
}

func newInstallRuntime(homeDir string, sel model.Selection, resolved planner.ResolvedPlan) (*installRuntime, error) {
	backupRoot := filepath.Join(homeDir, ".aispace-setup", "backups")
	if err := os.MkdirAll(backupRoot, 0o755); err != nil {
		return nil, fmt.Errorf("create backup root %q: %w", backupRoot, err)
	}
	return &installRuntime{
		homeDir:    homeDir,
		selection:  sel,
		resolved:   resolved,
		backupRoot: backupRoot,
		state:      &runtimeState{},
	}, nil
}

func (r *installRuntime) stagePlan() pipeline.StagePlan {
	adapters := resolveAdapters(r.selection.Agents)
	targets := componentBackupTargets(r.homeDir, r.selection, adapters, r.selection.Components)

	prepare := []pipeline.Step{
		prepareBackupStep{
			id:          "prepare:backup-snapshot",
			snapshotter: backup.NewSnapshotter(),
			snapshotDir: filepath.Join(r.backupRoot, time.Now().UTC().Format("20060102150405.000000000")),
			targets:     targets,
			state:       r.state,
			source:      backup.BackupSourceInstall,
			description: "pre-install snapshot",
			appVersion:  AppVersion,
		},
	}

	apply := make([]pipeline.Step, 0, 1+len(r.selection.Agents)+len(r.resolved.OrderedComponents))
	apply = append(apply, rollbackRestoreStep{id: "apply:rollback-restore", state: r.state})

	for _, agentID := range r.selection.Agents {
		apply = append(apply, agentInstallStep{
			id:      "agent:" + string(agentID),
			agent:   agentID,
			homeDir: r.homeDir,
		})
	}

	for _, component := range r.resolved.OrderedComponents {
		apply = append(apply, componentApplyStep{
			id:        "component:" + string(component),
			component: component,
			homeDir:   r.homeDir,
			agents:    r.selection.Agents,
			selection: r.selection,
		})
	}

	return pipeline.StagePlan{Prepare: prepare, Apply: apply}
}

// — Step types —

type prepareBackupStep struct {
	id          string
	snapshotter backup.Snapshotter
	snapshotDir string
	targets     []string
	state       *runtimeState
	source      backup.BackupSource
	description string
	appVersion  string
}

func (s prepareBackupStep) ID() string { return s.id }

func (s prepareBackupStep) Run() error {
	manifest, err := s.snapshotter.Create(s.snapshotDir, s.targets)
	if err != nil {
		return fmt.Errorf("create backup snapshot: %w", err)
	}
	manifest.Source = s.source
	manifest.Description = s.description
	manifest.CreatedByVersion = s.appVersion
	manifestPath := filepath.Join(s.snapshotDir, backup.ManifestFilename)
	_ = backup.WriteManifest(manifestPath, manifest)
	s.state.manifest = manifest
	return nil
}

type rollbackRestoreStep struct {
	id    string
	state *runtimeState
}

func (s rollbackRestoreStep) ID() string   { return s.id }
func (s rollbackRestoreStep) Run() error   { return nil }
func (s rollbackRestoreStep) Rollback() error {
	if len(s.state.manifest.Entries) == 0 {
		return nil
	}
	return backup.RestoreService{}.Restore(s.state.manifest)
}

type agentInstallStep struct {
	id      string
	agent   model.AgentID
	homeDir string
}

func (s agentInstallStep) ID() string { return s.id }

func (s agentInstallStep) Run() error {
	adapter, err := agents.NewAdapter(s.agent)
	if err != nil {
		return fmt.Errorf("create adapter for %q: %w", s.agent, err)
	}
	if !adapter.SupportsAutoInstall() {
		return nil
	}
	installed, _, _, _, err := adapter.Detect(context.Background(), s.homeDir)
	if err != nil {
		return fmt.Errorf("detect agent %q: %w", s.agent, err)
	}
	if installed {
		return nil
	}
	commands, err := adapter.InstallCommand(system.PlatformProfile{})
	if err != nil {
		return fmt.Errorf("resolve install command for %q: %w", s.agent, err)
	}
	return runCommandSequence(commands)
}

type componentApplyStep struct {
	id        string
	component model.ComponentID
	homeDir   string
	agents    []model.AgentID
	selection model.Selection
}

func (s componentApplyStep) ID() string { return s.id }

func (s componentApplyStep) Run() error {
	adapters := resolveAdapters(s.agents)

	switch s.component {
	case model.ComponentEngram:
		commands, err := engram.InstallCommand(system.PlatformProfile{OS: "darwin", PackageManager: "brew"})
		if err == nil {
			_ = runCommandSequence(commands) // best-effort
		}
		for _, adapter := range adapters {
			if _, err := engram.Inject(s.homeDir, adapter); err != nil {
				return fmt.Errorf("inject engram for %q: %w", adapter.Agent(), err)
			}
		}
		return nil

	case model.ComponentSDD:
		for _, adapter := range adapters {
			if _, err := sdd.Inject(s.homeDir, adapter, s.selection.ClaudeModelAssigns); err != nil {
				return fmt.Errorf("inject sdd for %q: %w", adapter.Agent(), err)
			}
		}
		return nil

	case model.ComponentContext7:
		for _, adapter := range adapters {
			if _, err := mcp.Inject(s.homeDir, adapter); err != nil {
				return fmt.Errorf("inject context7 for %q: %w", adapter.Agent(), err)
			}
		}
		return nil

	case model.ComponentSkills:
		skillIDs := skills.SkillsForPreset(s.selection.Preset)
		if len(skillIDs) == 0 {
			return nil
		}
		for _, adapter := range adapters {
			if _, err := skills.Inject(s.homeDir, adapter, skillIDs); err != nil {
				return fmt.Errorf("inject skills for %q: %w", adapter.Agent(), err)
			}
		}
		return nil

	case model.ComponentPersona:
		for _, adapter := range adapters {
			if _, err := persona.Inject(s.homeDir, adapter, s.selection.Persona); err != nil {
				return fmt.Errorf("inject persona for %q: %w", adapter.Agent(), err)
			}
		}
		return nil

	case model.ComponentPermission:
		for _, adapter := range adapters {
			if _, err := permissions.Inject(s.homeDir, adapter); err != nil {
				return fmt.Errorf("inject permissions for %q: %w", adapter.Agent(), err)
			}
		}
		return nil

	case model.ComponentTheme:
		for _, adapter := range adapters {
			if _, err := theme.Inject(s.homeDir, adapter); err != nil {
				return fmt.Errorf("inject theme for %q: %w", adapter.Agent(), err)
			}
		}
		return nil

	case model.ComponentAISpace:
		return nil // placeholder

	default:
		return fmt.Errorf("unknown component %q", s.component)
	}
}

// — Helpers —

func resolveAdapters(agentIDs []model.AgentID) []agents.Adapter {
	result := make([]agents.Adapter, 0, len(agentIDs))
	for _, id := range agentIDs {
		adapter, err := agents.NewAdapter(id)
		if err != nil {
			continue
		}
		result = append(result, adapter)
	}
	return result
}

func componentPaths(homeDir string, sel model.Selection, adapters []agents.Adapter, component model.ComponentID) []string {
	var paths []string
	switch component {
	case model.ComponentPersona, model.ComponentEngram, model.ComponentSDD:
		for _, a := range adapters {
			paths = append(paths, a.SystemPromptFile(homeDir))
		}
	case model.ComponentContext7:
		for _, a := range adapters {
			paths = append(paths, a.MCPConfigPath(homeDir, "context7"))
			paths = append(paths, a.MCPConfigPath(homeDir, "aispace"))
		}
	case model.ComponentSkills:
		for _, a := range adapters {
			paths = append(paths, a.SkillsDir(homeDir))
		}
	case model.ComponentPermission, model.ComponentTheme:
		for _, a := range adapters {
			paths = append(paths, a.SettingsPath(homeDir))
		}
	}
	return paths
}

func componentBackupTargets(homeDir string, sel model.Selection, adapters []agents.Adapter, components []model.ComponentID) []string {
	seen := map[string]struct{}{}
	var targets []string
	for _, c := range components {
		for _, p := range componentPaths(homeDir, sel, adapters, c) {
			if _, ok := seen[p]; ok {
				continue
			}
			seen[p] = struct{}{}
			targets = append(targets, p)
		}
	}
	return targets
}

func discoverAgents(homeDir string) []model.AgentID {
	s, err := state.Read(homeDir)
	if err == nil && len(s.InstalledAgents) > 0 {
		return s.InstalledAgents
	}
	reg, err := agents.NewMVPRegistry()
	if err != nil {
		return nil
	}
	installed := agents.DiscoverInstalled(reg, homeDir)
	ids := make([]model.AgentID, 0, len(installed))
	for _, a := range installed {
		ids = append(ids, model.AgentID(a.ID))
	}
	return ids
}

func runCommandSequence(commands [][]string) error {
	for _, args := range commands {
		if len(args) == 0 {
			continue
		}
		cmd := exec.Command(args[0], args[1:]...) //nolint:gosec
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command %q failed: %w", args[0], err)
		}
	}
	return nil
}

func runPostInstallVerification(homeDir string, sel model.Selection, resolved planner.ResolvedPlan) []verify.CheckResult {
	adapters := resolveAdapters(sel.Agents)
	checks := make([]verify.Check, 0)

	for _, component := range resolved.OrderedComponents {
		for _, path := range componentPaths(homeDir, sel, adapters, component) {
			p := path
			checks = append(checks, verify.Check{
				ID:          "verify:install:file:" + p,
				Description: "installed file exists: " + p,
				Soft:        true,
				Run: func(_ context.Context) error {
					_, err := os.Stat(p)
					return err
				},
			})
		}
	}

	return verify.RunChecks(context.Background(), checks)
}

// ioDiscard silences flag error output.
type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }

// csvListFlag supports comma-separated flag values.
type csvListFlag struct{ values *[]string }

func (f csvListFlag) String() string {
	if f.values == nil {
		return ""
	}
	parts := make([]string, len(*f.values))
	copy(parts, *f.values)
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ","
		}
		result += p
	}
	return result
}

func (f csvListFlag) Set(value string) error {
	for _, part := range splitCSV(value) {
		*f.values = append(*f.values, part)
	}
	return nil
}

func registerListFlag(fs *flag.FlagSet, name string, values *[]string) {
	fs.Var(csvListFlag{values: values}, name, "comma-separated list")
}

func splitCSV(s string) []string {
	var parts []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			p := strings.TrimSpace(s[start:i])
			if p != "" {
				parts = append(parts, p)
			}
			start = i + 1
		}
	}
	return parts
}

