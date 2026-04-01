package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ismartz/aispace-setup/internal/model"
	"github.com/ismartz/aispace-setup/internal/pipeline"
	"github.com/ismartz/aispace-setup/internal/planner"
	"github.com/ismartz/aispace-setup/internal/system"
	"github.com/ismartz/aispace-setup/internal/tui/screens"
)

// TickMsg drives the spinner animation on the progress screen.
type TickMsg time.Time

// PipelineDoneMsg is sent when the install pipeline finishes execution.
type PipelineDoneMsg struct {
	Result pipeline.ExecutionResult
}

// SyncDoneMsg is sent when the sync operation completes.
type SyncDoneMsg struct {
	FilesChanged int
	Err          error
}

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Screen identifies which TUI screen is currently active.
type Screen int

const (
	ScreenWelcome     Screen = iota
	ScreenAgents
	ScreenPersona
	ScreenPreset
	ScreenModelPicker // shown when SDD component is included
	ScreenReview
	ScreenProgress
	ScreenResult
)

// SyncFunc is the function the TUI calls to perform a sync (bound to homeDir by app.go).
type SyncFunc func() (int, error)

// ExecuteFunc builds and runs the install pipeline.
type ExecuteFunc func(
	selection model.Selection,
	resolved planner.ResolvedPlan,
	detection system.PlatformProfile,
	onProgress pipeline.ProgressFunc,
) pipeline.ExecutionResult

// Model is the root Bubbletea model.
type Model struct {
	screen    Screen
	detection system.PlatformProfile
	version   string

	// Selection state (built up across screens)
	selectedAgents       []model.AgentID
	selectedPersona      model.PersonaID
	selectedPreset       model.PresetID
	modelPickerState     screens.ModelPickerState

	// Resolved plan (set when ScreenReview starts the pipeline)
	resolved planner.ResolvedPlan
	plan     pipeline.StagePlan

	// Progress tracking
	stepStatuses map[string]pipeline.StepStatus
	currentStep  string
	tick         int

	// Result state
	installResult pipeline.ExecutionResult
	syncResult    struct {
		filesChanged int
		err          error
	}
	isSyncMode bool

	// Injected functions (set by app.go)
	ExecuteFn ExecuteFunc
	SyncFn    SyncFunc

	// Per-screen cursor positions
	cursors map[Screen]int

	err error
}

// NewModel constructs the initial TUI Model.
func NewModel(detection system.PlatformProfile, version string) Model {
	return Model{
		screen:          ScreenWelcome,
		detection:       detection,
		version:         version,
		selectedPersona: model.PersonaNeutral,
		selectedPreset:  model.PresetFull,
		stepStatuses:    make(map[string]pipeline.StepStatus),
		modelPickerState: screens.NewModelPickerState(),
		cursors: map[Screen]int{
			ScreenWelcome:     0,
			ScreenAgents:      0,
			ScreenPersona:     0,
			ScreenPreset:      0,
			ScreenModelPicker: 0,
			ScreenReview:      0,
			ScreenProgress:    0,
			ScreenResult:      0,
		},
	}
}

// Init starts the tick command for spinner animation.
func (m Model) Init() tea.Cmd {
	return tickCmd()
}
