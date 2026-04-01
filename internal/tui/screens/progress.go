package screens

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ismartz/aispace-setup/internal/pipeline"
	"github.com/ismartz/aispace-setup/internal/tui/styles"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// RenderProgress renders the progress screen.
func RenderProgress(tick int, currentStep string, stepStatuses map[string]pipeline.StepStatus, isSyncMode bool) string {
	var b strings.Builder

	title := "Installing..."
	if isSyncMode {
		title = "Syncing..."
	}
	b.WriteString(styles.HeadingStyle.Render(title))
	b.WriteString("\n\n")

	frame := spinnerFrames[tick%len(spinnerFrames)]
	spinnerStyle := lipgloss.NewStyle().Foreground(styles.AccentBlue).Bold(true)
	b.WriteString(spinnerStyle.Render(frame + " Running pipeline..."))
	b.WriteString("\n")

	if len(stepStatuses) > 0 {
		b.WriteString("\n")
		for stepID, status := range stepStatuses {
			icon := "·"
			iconStyle := styles.MutedStyle
			switch status {
			case pipeline.StepStatusRunning:
				icon = spinnerFrames[tick%len(spinnerFrames)]
				iconStyle = lipgloss.NewStyle().Foreground(styles.AccentBlue).Bold(true)
			case pipeline.StepStatusSucceeded:
				icon = "✓"
				iconStyle = styles.SuccessStyle
			case pipeline.StepStatusFailed:
				icon = "✗"
				iconStyle = styles.ErrorStyle
			}
			b.WriteString(iconStyle.Render("  "+icon+" "+stepID) + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(styles.MutedStyle.Render("Please wait..."))

	return styles.FrameStyle.Render(b.String())
}
