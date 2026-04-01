package screens

import (
	"strings"

	"github.com/ismartz/aispace-setup/internal/model"
	"github.com/ismartz/aispace-setup/internal/tui/styles"
)

type presetOption struct {
	id          model.PresetID
	label       string
	description string
	components  string
}

var presetOptions = []presetOption{
	{
		id:          model.PresetFull,
		label:       "Full",
		description: "Complete setup — everything enabled",
		components:  "engram, sdd, skills, context7, persona, permissions, theme",
	},
	{
		id:          model.PresetCore,
		label:       "Core",
		description: "Essentials only — no theme/permissions",
		components:  "engram, sdd, skills, context7",
	},
	{
		id:          model.PresetMinimal,
		label:       "Minimal",
		description: "Memory only — Engram",
		components:  "engram",
	},
	{
		id:          model.PresetCustom,
		label:       "Custom",
		description: "Choose components individually (advanced)",
		components:  "you select",
	},
}

// RenderPreset renders the preset selection screen.
func RenderPreset(cursor int, current model.PresetID) string {
	var b strings.Builder

	b.WriteString(styles.HeadingStyle.Render("Select Preset"))
	b.WriteString("\n\n")

	for i, opt := range presetOptions {
		selected := opt.id == current
		radio := renderRadio(opt.label, selected, i == cursor)
		desc := styles.MutedStyle.Render("   " + opt.description)
		b.WriteString(radio + "\n" + desc + "\n")
	}

	b.WriteString("\n")

	// Show components for highlighted preset
	highlighted := presetOptions[cursor]
	b.WriteString(styles.SubtextStyle.Render("Components: "))
	b.WriteString(styles.AccentStyle.Render(highlighted.components))
	b.WriteString("\n\n")
	b.WriteString(styles.HelpStyle.Render("↑↓ navigate  •  enter confirm  •  esc back  •  q quit"))

	return styles.FrameStyle.Render(b.String())
}
