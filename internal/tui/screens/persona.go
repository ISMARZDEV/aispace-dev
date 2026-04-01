package screens

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ismartz/aispace-setup/internal/model"
	"github.com/ismartz/aispace-setup/internal/tui/styles"
)

type personaOption struct {
	id      model.PersonaID
	label   string
	badge   string
	preview string
}

var personaOptions = []personaOption{
	{
		id:      model.PersonaNeutral,
		label:   "Neutral",
		badge:   "",
		preview: `"He identificado el problema en auth/middleware.go:47.` + "\n" + ` El token no se valida antes del acceso."`,
	},
	{
		id:      model.PersonaDominicano,
		label:   "Dominicano",
		badge:   "JARVIS-RD",
		preview: `"Jefe, encontré el problema — línea 47. El token no` + "\n" + ` se valida. Eso ta' fácil, lo arreglo."`,
	},
	{
		id:      model.PersonaAlien,
		label:   "Alien Observer",
		badge:   "OBSERVER",
		preview: `"Anomalía detectada. auth/middleware.go:47.` + "\n" + ` Probabilidad de fix: 99.7%."`,
	},
}

// RenderPersona renders the persona selection screen with live preview.
func RenderPersona(cursor int, current model.PersonaID) string {
	var b strings.Builder

	b.WriteString(styles.HeadingStyle.Render("Select Persona"))
	b.WriteString("\n\n")

	for i, opt := range personaOptions {
		line := renderRadio(opt.label, opt.id == current, i == cursor)
		if opt.badge != "" {
			badge := styles.PersonaBadge.Render(opt.badge)
			line = line + "  " + badge
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Preview box for currently highlighted persona
	highlighted := personaOptions[cursor]
	previewStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Muted).
		Padding(0, 1).
		Foreground(styles.Foreground)

	b.WriteString(styles.MutedStyle.Render("┄ Preview"))
	b.WriteString("\n")
	b.WriteString(previewStyle.Render(highlighted.preview))
	b.WriteString("\n\n")
	b.WriteString(styles.HelpStyle.Render("↑↓ navigate  •  enter confirm  •  esc back  •  q quit"))

	return styles.FrameStyle.Render(b.String())
}
