package screens

import (
	"strings"

	"github.com/ismartz/aispace-setup/internal/tui/styles"
)

// WelcomeMenuOptions are the top-level menu items.
var WelcomeMenuOptions = []string{
	"Install / Configure",
	"Sync configs",
	"Quit",
}

// RenderWelcome renders the welcome screen.
func RenderWelcome(cursor int, version string) string {
	var b strings.Builder

	b.WriteString(styles.RenderLogo())
	b.WriteString("\n\n")
	b.WriteString(styles.SubtextStyle.Render(styles.Tagline(version)))
	b.WriteString("\n\n")
	b.WriteString(renderSeparator())
	b.WriteString("\n\n")
	b.WriteString(styles.HeadingStyle.Render("Menu"))
	b.WriteString("\n\n")
	b.WriteString(renderOptions(WelcomeMenuOptions, cursor))
	b.WriteString("\n\n")
	b.WriteString(styles.HelpStyle.Render("j/k navigate  •  enter select  •  q quit"))

	return styles.FrameStyle.Render(b.String())
}
