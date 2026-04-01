package tui

import "github.com/ismartz/aispace-setup/internal/tui/screens"

// View renders the current screen.
func (m Model) View() string {
	switch m.screen {
	case ScreenWelcome:
		return screens.RenderWelcome(m.cursors[ScreenWelcome], m.version)
	case ScreenAgents:
		return screens.RenderAgents(m.cursors[ScreenAgents], m.selectedAgents)
	case ScreenPersona:
		return screens.RenderPersona(m.cursors[ScreenPersona], m.selectedPersona)
	case ScreenPreset:
		return screens.RenderPreset(m.cursors[ScreenPreset], m.selectedPreset)
	case ScreenModelPicker:
		return screens.RenderModelPicker(m.modelPickerState)
	case ScreenReview:
		return screens.RenderReview(m.selectedAgents, m.selectedPersona, m.selectedPreset, m.resolved)
	case ScreenProgress:
		return screens.RenderProgress(m.tick, m.currentStep, m.stepStatuses, m.isSyncMode)
	case ScreenResult:
		return screens.RenderResult(m.installResult, m.syncResult.filesChanged, m.syncResult.err, m.isSyncMode)
	}
	return ""
}
