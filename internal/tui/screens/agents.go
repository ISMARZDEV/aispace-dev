package screens

import (
	"strings"

	"github.com/ismartz/aispace-setup/internal/model"
	"github.com/ismartz/aispace-setup/internal/tui/styles"
)

// availableAgents is the ordered list shown on the agents screen.
var availableAgents = []model.AgentID{
	model.AgentClaudeCode,
	model.AgentOpenCode,
}

func agentLabel(id model.AgentID) string {
	switch id {
	case model.AgentClaudeCode:
		return "Claude Code"
	case model.AgentOpenCode:
		return "OpenCode"
	default:
		return string(id)
	}
}

func isAgentSelected(selected []model.AgentID, id model.AgentID) bool {
	for _, a := range selected {
		if a == id {
			return true
		}
	}
	return false
}

// RenderAgents renders the agent multi-select screen.
func RenderAgents(cursor int, selectedAgents []model.AgentID) string {
	var b strings.Builder

	b.WriteString(styles.HeadingStyle.Render("Select Agents"))
	b.WriteString("\n\n")

	for i, agent := range availableAgents {
		checked := isAgentSelected(selectedAgents, agent)
		b.WriteString(renderCheckbox(agentLabel(agent), checked, i == cursor))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if len(selectedAgents) == 0 {
		b.WriteString(styles.WarningStyle.Render("Select at least one agent to continue"))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("space toggle  •  enter confirm  •  esc back  •  q quit"))

	return styles.FrameStyle.Render(b.String())
}
