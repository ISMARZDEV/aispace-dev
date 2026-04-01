package app

import (
	"context"
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ismartz/aispace-setup/internal/cli"
	"github.com/ismartz/aispace-setup/internal/model"
	"github.com/ismartz/aispace-setup/internal/pipeline"
	"github.com/ismartz/aispace-setup/internal/planner"
	"github.com/ismartz/aispace-setup/internal/system"
	"github.com/ismartz/aispace-setup/internal/tui"
	"github.com/ismartz/aispace-setup/internal/verify"
)

// Run is the main entry point.
func Run() error {
	return RunArgs(os.Args[1:], os.Stdout)
}

// RunArgs dispatches to TUI or CLI based on args.
func RunArgs(args []string, stdout io.Writer) error {
	cli.AppVersion = Version

	profile, err := system.Detect(context.Background())
	if err != nil {
		return fmt.Errorf("detect system: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}

	if len(args) == 0 {
		m := tui.NewModel(profile, Version)
		m.ExecuteFn = tuiExecute
		m.SyncFn = tuiSync(homeDir)
		p := tea.NewProgram(m, tea.WithAltScreen())
		_, err := p.Run()
		return err
	}

	switch args[0] {
	case "version", "--version", "-v":
		_, _ = fmt.Fprintf(stdout, "ai-setup %s\n", Version)
		return nil

	case "install":
		result, err := cli.RunInstall(args[1:], profile)
		if err != nil {
			return err
		}
		if result.DryRun {
			_, _ = fmt.Fprintln(stdout, cli.RenderDryRun(result))
			return nil
		}
		_, _ = fmt.Fprintln(stdout, renderVerifyReport(result.Verify))
		return nil

	case "sync":
		result, err := cli.RunSync(args[1:])
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(stdout, cli.RenderSyncReport(result))
		return nil

	default:
		return fmt.Errorf("unknown command %q — try: install, sync, version", args[0])
	}
}

// tuiExecute is the TUI's install callback.
func tuiExecute(
	selection model.Selection,
	resolved planner.ResolvedPlan,
	detection system.PlatformProfile,
	onProgress pipeline.ProgressFunc,
) pipeline.ExecutionResult {
	return cli.ExecuteInstallPlan(selection, resolved, detection, onProgress)
}

// tuiSync returns the TUI's sync callback bound to homeDir.
func tuiSync(homeDir string) tui.SyncFunc {
	return func() (int, error) {
		agentIDs := cli.DiscoverAgents(homeDir)
		sel := cli.BuildSyncSelection(cli.SyncFlags{}, agentIDs)
		result, err := cli.RunSyncWithSelection(homeDir, sel)
		if err != nil {
			return 0, err
		}
		return result.FilesChanged, nil
	}
}

// renderVerifyReport renders a minimal post-install summary.
func renderVerifyReport(results []verify.CheckResult) string {
	passed, warnings, failed := 0, 0, 0
	for _, r := range results {
		switch r.Status {
		case verify.CheckStatusPassed:
			passed++
		case verify.CheckStatusFailed:
			failed++
		case verify.CheckStatusWarning:
			warnings++
		}
	}
	if failed == 0 {
		return fmt.Sprintf("ai-setup install — complete\nChecks: %d passed, %d warnings", passed, warnings)
	}
	return fmt.Sprintf("ai-setup install — complete with issues\nChecks: %d passed, %d warnings, %d failed", passed, warnings, failed)
}
