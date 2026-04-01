package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ismartz/aispace-setup/internal/agents"
	"github.com/ismartz/aispace-setup/internal/backup"
	"github.com/ismartz/aispace-setup/internal/components/engram"
	"github.com/ismartz/aispace-setup/internal/components/mcp"
	"github.com/ismartz/aispace-setup/internal/components/permissions"
	"github.com/ismartz/aispace-setup/internal/components/sdd"
	"github.com/ismartz/aispace-setup/internal/components/skills"
	"github.com/ismartz/aispace-setup/internal/components/theme"
	"github.com/ismartz/aispace-setup/internal/model"
	"github.com/ismartz/aispace-setup/internal/pipeline"
	"github.com/ismartz/aispace-setup/internal/state"
	"github.com/ismartz/aispace-setup/internal/verify"
)

// SyncFlags holds parsed CLI flags for the sync command.
type SyncFlags struct {
	Agents             []string
	SDDMode            string
	StrictTDD          bool
	IncludePermissions bool
	IncludeTheme       bool
	DryRun             bool
}

// SyncResult holds the outcome of a sync execution.
type SyncResult struct {
	Agents       []model.AgentID
	Selection    model.Selection
	Plan         pipeline.StagePlan
	Execution    pipeline.ExecutionResult
	Verify       []verify.CheckResult
	DryRun       bool
	NoOp         bool
	FilesChanged int
}

// ParseSyncFlags parses CLI args for the sync subcommand.
func ParseSyncFlags(args []string) (SyncFlags, error) {
	var opts SyncFlags

	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	fs.SetOutput(ioDiscard{})
	registerListFlag(fs, "agent", &opts.Agents)
	registerListFlag(fs, "agents", &opts.Agents)
	fs.StringVar(&opts.SDDMode, "sdd-mode", "", "SDD orchestrator mode: single or multi (default: single)")
	fs.BoolVar(&opts.StrictTDD, "strict-tdd", false, "enable strict TDD mode for SDD agents")
	fs.BoolVar(&opts.IncludePermissions, "include-permissions", false, "include permissions in sync")
	fs.BoolVar(&opts.IncludeTheme, "include-theme", false, "include theme in sync")
	fs.BoolVar(&opts.DryRun, "dry-run", false, "preview plan without executing")

	if err := fs.Parse(args); err != nil {
		return SyncFlags{}, err
	}

	if fs.NArg() > 0 {
		return SyncFlags{}, fmt.Errorf("unexpected sync argument %q", fs.Arg(0))
	}

	return opts, nil
}

// BuildSyncSelection builds model.Selection for sync.
// Default scope: SDD, Engram, Context7, Skills.
// Persona and AISpace excluded by default (user-config-adjacent).
func BuildSyncSelection(flags SyncFlags, agentIDs []model.AgentID) model.Selection {
	components := []model.ComponentID{
		model.ComponentSDD,
		model.ComponentEngram,
		model.ComponentContext7,
		model.ComponentSkills,
	}
	if flags.IncludePermissions {
		components = append(components, model.ComponentPermission)
	}
	if flags.IncludeTheme {
		components = append(components, model.ComponentTheme)
	}

	sddMode := model.SDDModeID(flags.SDDMode)
	if sddMode == "" {
		sddMode = model.SDDModeSingle
	}

	return model.Selection{
		Agents:     agentIDs,
		Components: components,
		SDDMode:    sddMode,
		StrictTDD:  flags.StrictTDD,
		Preset:     model.PresetFull,
	}
}

// DiscoverAgents returns agent IDs to sync.
// Reads persisted state first; falls back to filesystem discovery.
func DiscoverAgents(homeDir string) []model.AgentID {
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

// syncRuntime mirrors installRuntime but only calls inject functions.
type syncRuntime struct {
	homeDir      string
	selection    model.Selection
	agentIDs     []model.AgentID
	backupRoot   string
	state        *runtimeState
	filesChanged int
}

func newSyncRuntime(homeDir string, selection model.Selection) (*syncRuntime, error) {
	backupRoot := filepath.Join(homeDir, ".aispace-setup", "backups")
	if err := os.MkdirAll(backupRoot, 0o755); err != nil {
		return nil, fmt.Errorf("create backup root %q: %w", backupRoot, err)
	}
	return &syncRuntime{
		homeDir:    homeDir,
		selection:  selection,
		agentIDs:   selection.Agents,
		backupRoot: backupRoot,
		state:      &runtimeState{},
	}, nil
}

func (r *syncRuntime) stagePlan() pipeline.StagePlan {
	adapters := resolveAdapters(r.agentIDs)
	targets := componentBackupTargets(r.homeDir, r.selection, adapters, r.selection.Components)

	prepare := []pipeline.Step{
		prepareBackupStep{
			id:          "prepare:backup-snapshot",
			snapshotter: backup.NewSnapshotter(),
			snapshotDir: filepath.Join(r.backupRoot, time.Now().UTC().Format("20060102150405.000000000")),
			targets:     targets,
			state:       r.state,
			source:      backup.BackupSourceSync,
			description: "pre-sync snapshot",
			appVersion:  AppVersion,
		},
	}

	apply := []pipeline.Step{
		rollbackRestoreStep{id: "apply:rollback-restore", state: r.state},
	}

	for _, component := range r.selection.Components {
		apply = append(apply, componentSyncStep{
			id:           "sync:" + string(component),
			component:    component,
			homeDir:      r.homeDir,
			agents:       r.agentIDs,
			selection:    r.selection,
			filesChanged: &r.filesChanged,
		})
	}

	return pipeline.StagePlan{Prepare: prepare, Apply: apply}
}

// componentSyncStep calls inject only — no binary install, no persona.
type componentSyncStep struct {
	id           string
	component    model.ComponentID
	homeDir      string
	agents       []model.AgentID
	selection    model.Selection
	filesChanged *int
}

func (s componentSyncStep) ID() string { return s.id }

func (s componentSyncStep) Run() error {
	adapters := resolveAdapters(s.agents)

	switch s.component {
	case model.ComponentEngram:
		for _, a := range adapters {
			res, err := engram.Inject(s.homeDir, a)
			if err != nil {
				return fmt.Errorf("sync engram for %q: %w", a.Agent(), err)
			}
			s.countChanged(boolToInt(res.Changed))
		}
		return nil

	case model.ComponentContext7:
		for _, a := range adapters {
			res, err := mcp.Inject(s.homeDir, a)
			if err != nil {
				return fmt.Errorf("sync context7 for %q: %w", a.Agent(), err)
			}
			s.countChanged(boolToInt(res.Changed))
		}
		return nil

	case model.ComponentSDD:
		for _, a := range adapters {
			res, err := sdd.Inject(s.homeDir, a, nil)
			if err != nil {
				return fmt.Errorf("sync sdd for %q: %w", a.Agent(), err)
			}
			s.countChanged(boolToInt(res.Changed))
		}
		return nil

	case model.ComponentSkills:
		skillIDs := skills.SkillsForPreset(s.selection.Preset)
		if len(skillIDs) == 0 {
			return nil
		}
		for _, a := range adapters {
			res, err := skills.Inject(s.homeDir, a, skillIDs)
			if err != nil {
				return fmt.Errorf("sync skills for %q: %w", a.Agent(), err)
			}
			s.countChanged(boolToInt(res.Changed))
		}
		return nil

	case model.ComponentPermission:
		for _, a := range adapters {
			res, err := permissions.Inject(s.homeDir, a)
			if err != nil {
				return fmt.Errorf("sync permissions for %q: %w", a.Agent(), err)
			}
			s.countChanged(boolToInt(res.Changed))
		}
		return nil

	case model.ComponentTheme:
		for _, a := range adapters {
			res, err := theme.Inject(s.homeDir, a)
			if err != nil {
				return fmt.Errorf("sync theme for %q: %w", a.Agent(), err)
			}
			s.countChanged(boolToInt(res.Changed))
		}
		return nil

	default:
		return fmt.Errorf("component %q not supported in sync", s.component)
	}
}

func (s componentSyncStep) countChanged(n int) {
	if s.filesChanged != nil && n > 0 {
		*s.filesChanged += n
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// RunSyncWithSelection is the programmatic entry point for sync (used by TUI).
func RunSyncWithSelection(homeDir string, selection model.Selection) (SyncResult, error) {
	result := SyncResult{Agents: selection.Agents, Selection: selection}

	if len(selection.Agents) == 0 {
		result.NoOp = true
		return result, nil
	}

	rt, err := newSyncRuntime(homeDir, selection)
	if err != nil {
		return result, err
	}

	stagePlan := rt.stagePlan()
	result.Plan = stagePlan

	result.Execution = pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy()).Execute(stagePlan)
	if result.Execution.Err != nil {
		return result, fmt.Errorf("execute sync pipeline: %w", result.Execution.Err)
	}

	result.FilesChanged = rt.filesChanged
	if result.FilesChanged == 0 {
		result.NoOp = true
	}

	result.Verify = runPostSyncVerification(homeDir, selection)
	if verify.AnyFailed(result.Verify) {
		return result, fmt.Errorf("post-sync verification failed")
	}

	return result, nil
}

// RunSync is the CLI entry point for sync.
func RunSync(args []string) (SyncResult, error) {
	flags, err := ParseSyncFlags(args)
	if err != nil {
		return SyncResult{}, err
	}

	homeDir, err := osUserHomeDir()
	if err != nil {
		return SyncResult{}, fmt.Errorf("resolve home directory: %w", err)
	}

	var agentIDs []model.AgentID
	if len(flags.Agents) > 0 {
		agentIDs = asAgentIDs(flags.Agents)
	} else {
		agentIDs = DiscoverAgents(homeDir)
	}
	agentIDs = unique(agentIDs)

	selection := BuildSyncSelection(flags, agentIDs)

	if flags.DryRun {
		result := SyncResult{Agents: agentIDs, Selection: selection, DryRun: true}
		if len(agentIDs) == 0 {
			result.NoOp = true
			return result, nil
		}
		rt, err := newSyncRuntime(homeDir, selection)
		if err != nil {
			return result, err
		}
		result.Plan = rt.stagePlan()
		return result, nil
	}

	result, err := RunSyncWithSelection(homeDir, selection)
	if err != nil {
		return result, err
	}
	result.DryRun = false
	return result, nil
}

// RenderSyncReport renders a human-readable sync summary.
func RenderSyncReport(result SyncResult) string {
	var b strings.Builder

	if result.NoOp {
		fmt.Fprintln(&b, "ai-setup sync — no managed sync actions needed")
		if len(result.Agents) == 0 {
			fmt.Fprintln(&b, "No agents discovered or specified. Nothing to sync.")
		} else {
			fmt.Fprintf(&b, "Agents: %s\n", joinAgentIDs(result.Agents))
			fmt.Fprintln(&b, "All managed assets already up to date. No files changed.")
		}
		return strings.TrimRight(b.String(), "\n")
	}

	if result.DryRun {
		fmt.Fprintln(&b, "ai-setup sync — dry-run")
		fmt.Fprintf(&b, "Agents: %s\n", joinAgentIDs(result.Agents))
		compParts := make([]string, 0, len(result.Selection.Components))
		for _, c := range result.Selection.Components {
			compParts = append(compParts, string(c))
		}
		if len(compParts) > 0 {
			fmt.Fprintf(&b, "Components: %s\n", strings.Join(compParts, ", "))
		}
		fmt.Fprintf(&b, "Prepare steps: %d\n", len(result.Plan.Prepare))
		fmt.Fprintf(&b, "Apply steps: %d\n", len(result.Plan.Apply))
		return strings.TrimRight(b.String(), "\n")
	}

	fmt.Fprintln(&b, "ai-setup sync — executed")
	fmt.Fprintf(&b, "Agents: %s\n", joinAgentIDs(result.Agents))
	fmt.Fprintf(&b, "Files changed: %d\n", result.FilesChanged)

	return strings.TrimRight(b.String(), "\n")
}

func runPostSyncVerification(homeDir string, sel model.Selection) []verify.CheckResult {
	adapters := resolveAdapters(sel.Agents)
	checks := make([]verify.Check, 0)
	for _, c := range sel.Components {
		for _, p := range componentPaths(homeDir, sel, adapters, c) {
			path := p
			checks = append(checks, verify.Check{
				ID:          "verify:sync:file:" + path,
				Description: "synced file exists: " + path,
				Soft:        true,
				Run: func(_ context.Context) error {
					_, err := os.Stat(path)
					return err
				},
			})
		}
	}
	return verify.RunChecks(context.Background(), checks)
}
