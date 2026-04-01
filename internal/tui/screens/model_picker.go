package screens

import (
	"fmt"
	"strings"

	"github.com/ismartz/aispace-setup/internal/model"
	"github.com/ismartz/aispace-setup/internal/tui/styles"
)

type modelPresetOption struct {
	id          model.ClaudeModelPreset
	label       string
	description string
}

var modelPresetOptions = []modelPresetOption{
	{
		id:          model.ClaudePresetBalanced,
		label:       "Balanced",
		description: "Opus for architecture & verify, Sonnet for most phases, Haiku for archive",
	},
	{
		id:          model.ClaudePresetPerformance,
		label:       "Performance",
		description: "Opus for all critical phases — highest quality, higher cost",
	},
	{
		id:          model.ClaudePresetEconomy,
		label:       "Economy",
		description: "Sonnet for all phases, Haiku for archive — cost optimised",
	},
	{
		id:          model.ClaudePresetCustom,
		label:       "Custom",
		description: "Pick the model for each SDD phase individually",
	},
}

var phaseLabels = map[model.SDDPhase]string{
	model.SDDPhaseOrchestrator: "Orchestrator (coordinator)",
	model.SDDPhaseInit:         "Init         (requirements)",
	model.SDDPhaseExplore:      "Explore      (codebase scan)",
	model.SDDPhasePropose:      "Propose      (approaches)",
	model.SDDPhaseSpec:         "Spec         (specification)",
	model.SDDPhaseDesign:       "Design       (interfaces)",
	model.SDDPhaseTasks:        "Tasks        (task breakdown)",
	model.SDDPhaseApply:        "Apply        (implementation)",
	model.SDDPhaseVerify:       "Verify       (validation)",
	model.SDDPhaseArchive:      "Archive      (documentation)",
}

var aliasOrder = []model.ClaudeModelAlias{
	model.ClaudeModelOpus,
	model.ClaudeModelSonnet,
	model.ClaudeModelHaiku,
}

// ModelPickerState holds the state for the model picker screen.
type ModelPickerState struct {
	Preset      model.ClaudeModelPreset
	Assignments model.ClaudeModelAssignments
	// In custom mode: cursor over phases; otherwise cursor over presets
	CustomMode  bool
	Cursor      int
}

// NewModelPickerState returns the initial state with the balanced preset.
func NewModelPickerState() ModelPickerState {
	return ModelPickerState{
		Preset:      model.ClaudePresetBalanced,
		Assignments: model.DefaultClaudeAssignments(model.ClaudePresetBalanced),
		CustomMode:  false,
		Cursor:      0,
	}
}

// SelectPreset applies a preset and updates assignments.
func (s *ModelPickerState) SelectPreset(preset model.ClaudeModelPreset) {
	s.Preset = preset
	if preset == model.ClaudePresetCustom {
		s.CustomMode = true
		s.Cursor = 0
	} else {
		s.Assignments = model.DefaultClaudeAssignments(preset)
		s.CustomMode = false
	}
}

// CyclePhaseModel cycles the model for the phase at the given index.
func (s *ModelPickerState) CyclePhaseModel(phaseIdx int) {
	phases := model.AllSDDPhases()
	if phaseIdx < 0 || phaseIdx >= len(phases) {
		return
	}
	phase := phases[phaseIdx]
	current := s.Assignments[phase]
	for i, alias := range aliasOrder {
		if alias == current {
			s.Assignments[phase] = aliasOrder[(i+1)%len(aliasOrder)]
			return
		}
	}
	s.Assignments[phase] = model.ClaudeModelSonnet
}

// RenderModelPicker renders the model assignment screen.
func RenderModelPicker(s ModelPickerState) string {
	if s.CustomMode {
		return renderCustomPicker(s)
	}
	return renderPresetPicker(s)
}

func renderPresetPicker(s ModelPickerState) string {
	var b strings.Builder

	b.WriteString(styles.HeadingStyle.Render("Assign Models to SDD Phases"))
	b.WriteString("\n\n")

	for i, opt := range modelPresetOptions {
		selected := opt.id == s.Preset
		radio := renderRadio(opt.label, selected, i == s.Cursor)
		desc := styles.MutedStyle.Render("   " + opt.description)
		b.WriteString(radio + "\n" + desc + "\n")
	}

	b.WriteString("\n")
	b.WriteString(renderAssignmentPreview(s.Assignments))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("↑↓ navigate  •  enter select  •  esc back  •  q quit"))

	return styles.FrameStyle.Render(b.String())
}

func renderCustomPicker(s ModelPickerState) string {
	var b strings.Builder

	b.WriteString(styles.HeadingStyle.Render("Assign Models to SDD Phases"))
	b.WriteString("\n")
	b.WriteString(styles.MutedStyle.Render("Custom — press enter to cycle: opus → sonnet → haiku"))
	b.WriteString("\n\n")

	phases := model.AllSDDPhases()
	for i, phase := range phases {
		alias := s.Assignments[phase]
		label := phaseLabels[phase]
		row := fmt.Sprintf("  %-38s %s", label, renderAliasBadge(alias))
		if i == s.Cursor {
			row = styles.SelectedStyle.Render("> " + row[2:])
		}
		b.WriteString(row + "\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("↑↓ navigate  •  enter cycle model  •  c confirm  •  esc back"))

	return styles.FrameStyle.Render(b.String())
}

func renderAssignmentPreview(assigns model.ClaudeModelAssignments) string {
	var b strings.Builder
	b.WriteString(styles.SubtextStyle.Render("Current assignments:"))
	b.WriteString("\n")

	phases := model.AllSDDPhases()
	for _, phase := range phases {
		alias := assigns[phase]
		label := phaseLabels[phase]
		b.WriteString(fmt.Sprintf("  %-38s %s\n", label, renderAliasBadge(alias)))
	}
	return b.String()
}

func renderAliasBadge(alias model.ClaudeModelAlias) string {
	switch alias {
	case model.ClaudeModelOpus:
		return styles.AccentStyle.Render("opus  ")
	case model.ClaudeModelHaiku:
		return styles.MutedStyle.Render("haiku ")
	default:
		return styles.SubtextStyle.Render("sonnet")
	}
}
