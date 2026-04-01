package screens

import (
	"fmt"
	"strings"

	"github.com/ismartz/aispace-setup/internal/model"
	"github.com/ismartz/aispace-setup/internal/planner"
	"github.com/ismartz/aispace-setup/internal/tui/styles"
)

// RenderReview renders the installation plan review screen.
func RenderReview(
	agents []model.AgentID,
	persona model.PersonaID,
	preset model.PresetID,
	resolved planner.ResolvedPlan,
) string {
	var b strings.Builder

	b.WriteString(styles.HeadingStyle.Render("Review Installation Plan"))
	b.WriteString("\n\n")

	// Agents
	agentLabels := make([]string, 0, len(agents))
	for _, a := range agents {
		agentLabels = append(agentLabels, agentLabel(a))
	}
	b.WriteString(styles.SubtextStyle.Render("Agents:   "))
	b.WriteString(styles.TitleStyle.Render(strings.Join(agentLabels, ", ")))
	b.WriteString("\n")

	b.WriteString(styles.SubtextStyle.Render("Persona:  "))
	b.WriteString(styles.TitleStyle.Render(string(persona)))
	b.WriteString("\n")

	b.WriteString(styles.SubtextStyle.Render("Preset:   "))
	b.WriteString(styles.TitleStyle.Render(string(preset)))
	b.WriteString("\n\n")

	// Installation order
	if len(resolved.OrderedComponents) > 0 {
		b.WriteString(styles.SubtextStyle.Render("Installation order:"))
		b.WriteString("\n")
		for i, c := range resolved.OrderedComponents {
			b.WriteString(styles.MutedStyle.Render(fmt.Sprintf("  %d. %s", i+1, string(c))))
			b.WriteString("\n")
		}
	}

	// Auto-added dependencies
	if len(resolved.AddedDependencies) > 0 {
		b.WriteString("\n")
		deps := make([]string, 0, len(resolved.AddedDependencies))
		for _, d := range resolved.AddedDependencies {
			deps = append(deps, string(d))
		}
		b.WriteString(styles.WarningStyle.Render("Auto-added: " + strings.Join(deps, ", ")))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("enter to install  •  esc back  •  q quit"))

	return styles.FrameStyle.Render(b.String())
}
