package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ismartz/aispace-setup/internal/model"
	"github.com/ismartz/aispace-setup/internal/pipeline"
	"github.com/ismartz/aispace-setup/internal/planner"
)

var allAgents = []model.AgentID{
	model.AgentClaudeCode,
	model.AgentOpenCode,
}

var allPersonas = []model.PersonaID{
	model.PersonaNeutral,
	model.PersonaDominicano,
	model.PersonaAlien,
}

var allPresets = []model.PresetID{
	model.PresetFull,
	model.PresetCore,
	model.PresetMinimal,
	model.PresetCustom,
}

// Update handles all Bubbletea messages and key events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		m.tick++
		return m, tickCmd()

	case PipelineDoneMsg:
		m.installResult = msg.Result
		m.screen = ScreenResult
		return m, nil

	case SyncDoneMsg:
		m.syncResult.filesChanged = msg.FilesChanged
		m.syncResult.err = msg.Err
		m.screen = ScreenResult
		return m, nil

	case tea.KeyMsg:
		switch m.screen {
		case ScreenWelcome:
			return m.updateWelcome(msg)
		case ScreenAgents:
			return m.updateAgents(msg)
		case ScreenPersona:
			return m.updatePersona(msg)
		case ScreenPreset:
			return m.updatePreset(msg)
		case ScreenModelPicker:
			return m.updateModelPicker(msg)
		case ScreenReview:
			return m.updateReview(msg)
		case ScreenProgress:
			return m.updateProgress(msg)
		case ScreenResult:
			return m.updateResult(msg)
		}
	}

	return m, nil
}

func (m Model) updateWelcome(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	const maxCursor = 2
	switch msg.String() {
	case "j", "s", "down":
		if m.cursors[ScreenWelcome] < maxCursor {
			m.cursors[ScreenWelcome]++
		}
	case "k", "w", "up":
		if m.cursors[ScreenWelcome] > 0 {
			m.cursors[ScreenWelcome]--
		}
	case "enter":
		switch m.cursors[ScreenWelcome] {
		case 0: // Install / Configure
			m.screen = ScreenAgents
		case 1: // Sync configs
			m.isSyncMode = true
			m.screen = ScreenProgress
			return m, m.startSync()
		case 2: // Quit
			return m, tea.Quit
		}
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) updateAgents(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxCursor := len(allAgents) - 1
	switch msg.String() {
	case "j", "s", "down":
		if m.cursors[ScreenAgents] < maxCursor {
			m.cursors[ScreenAgents]++
		}
	case "k", "w", "up":
		if m.cursors[ScreenAgents] > 0 {
			m.cursors[ScreenAgents]--
		}
	case " ":
		agent := allAgents[m.cursors[ScreenAgents]]
		m.selectedAgents = toggleAgent(m.selectedAgents, agent)
	case "enter":
		if len(m.selectedAgents) > 0 {
			m.screen = ScreenPersona
		}
	case "esc", "v":
		m.screen = ScreenWelcome
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) updatePersona(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxCursor := len(allPersonas) - 1
	switch msg.String() {
	case "j", "s", "down":
		if m.cursors[ScreenPersona] < maxCursor {
			m.cursors[ScreenPersona]++
		}
	case "k", "w", "up":
		if m.cursors[ScreenPersona] > 0 {
			m.cursors[ScreenPersona]--
		}
	case "enter":
		m.selectedPersona = allPersonas[m.cursors[ScreenPersona]]
		m.screen = ScreenPreset
	case "esc", "v":
		m.screen = ScreenAgents
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) updatePreset(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxCursor := len(allPresets) - 1
	switch msg.String() {
	case "j", "s", "down":
		if m.cursors[ScreenPreset] < maxCursor {
			m.cursors[ScreenPreset]++
		}
	case "k", "w", "up":
		if m.cursors[ScreenPreset] > 0 {
			m.cursors[ScreenPreset]--
		}
	case "enter":
		m.selectedPreset = allPresets[m.cursors[ScreenPreset]]
		// Show model picker if SDD is included in the preset.
		if presetIncludesSDD(m.selectedPreset) {
			m.screen = ScreenModelPicker
		} else {
			sel := m.buildSelection()
			resolved, err := planner.NewResolver(planner.DefaultGraph()).Resolve(sel)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.resolved = resolved
			m.screen = ScreenReview
		}
	case "esc", "v":
		m.screen = ScreenPersona
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) updateReview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.screen = ScreenProgress
		return m, m.startInstall()
	case "esc", "v":
		m.screen = ScreenPreset
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) updateProgress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) updateResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "enter":
		return m, tea.Quit
	case "esc", "v":
		fresh := NewModel(m.detection, m.version)
		fresh.ExecuteFn = m.ExecuteFn
		fresh.SyncFn = m.SyncFn
		return fresh, tickCmd()
	}
	return m, nil
}

// startInstall launches the install pipeline in a goroutine.
func (m Model) startInstall() tea.Cmd {
	return func() tea.Msg {
		sel := m.buildSelection()
		onProgress := func(event pipeline.ProgressEvent) {}
		result := m.ExecuteFn(sel, m.resolved, m.detection, onProgress)
		return PipelineDoneMsg{Result: result}
	}
}

// startSync launches the sync in a goroutine.
func (m Model) startSync() tea.Cmd {
	return func() tea.Msg {
		filesChanged, err := m.SyncFn()
		return SyncDoneMsg{FilesChanged: filesChanged, Err: err}
	}
}

func (m Model) updateModelPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	s := m.modelPickerState

	if s.CustomMode {
		maxCursor := len(model.AllSDDPhases()) - 1
		switch msg.String() {
		case "j", "s", "down":
			if s.Cursor < maxCursor {
				s.Cursor++
			}
		case "k", "w", "up":
			if s.Cursor > 0 {
				s.Cursor--
			}
		case " ":
			s.CyclePhaseModel(s.Cursor)
		case "enter":
			s.CyclePhaseModel(s.Cursor)
		case "c":
			m.modelPickerState = s
			return m.resolveAndAdvance()
		case "esc", "v":
			s.CustomMode = false
			s.Cursor = 3 // back to "Custom" row in preset list
		case "q":
			return m, tea.Quit
		}
		m.modelPickerState = s
		return m, nil
	}

	maxCursor := 3 // 4 preset options
	switch msg.String() {
	case "j", "s", "down":
		if s.Cursor < maxCursor {
			s.Cursor++
		}
	case "k", "w", "up":
		if s.Cursor > 0 {
			s.Cursor--
		}
	case "enter":
		presets := []model.ClaudeModelPreset{
			model.ClaudePresetBalanced,
			model.ClaudePresetPerformance,
			model.ClaudePresetEconomy,
			model.ClaudePresetCustom,
		}
		s.SelectPreset(presets[s.Cursor])
		m.modelPickerState = s
		if s.CustomMode {
			return m, nil
		}
		return m.resolveAndAdvance()
	case "esc", "v":
		m.modelPickerState = s
		m.screen = ScreenPreset
	case "q":
		return m, tea.Quit
	}
	m.modelPickerState = s
	return m, nil
}

func (m Model) resolveAndAdvance() (tea.Model, tea.Cmd) {
	sel := m.buildSelection()
	resolved, err := planner.NewResolver(planner.DefaultGraph()).Resolve(sel)
	if err != nil {
		m.err = err
		return m, nil
	}
	m.resolved = resolved
	m.screen = ScreenReview
	return m, nil
}

// presetIncludesSDD returns true when the preset includes the SDD component.
func presetIncludesSDD(preset model.PresetID) bool {
	switch preset {
	case model.PresetFull, model.PresetCore:
		return true
	}
	return false
}

// buildSelection constructs a model.Selection from current TUI state.
func (m Model) buildSelection() model.Selection {
	return model.Selection{
		Agents:             m.selectedAgents,
		Persona:            m.selectedPersona,
		Preset:             m.selectedPreset,
		ClaudeModelPreset:  m.modelPickerState.Preset,
		ClaudeModelAssigns: m.modelPickerState.Assignments,
	}
}

// toggleAgent adds agent if absent, removes it if present.
func toggleAgent(agents []model.AgentID, agent model.AgentID) []model.AgentID {
	for i, a := range agents {
		if a == agent {
			return append(agents[:i], agents[i+1:]...)
		}
	}
	return append(agents, agent)
}
