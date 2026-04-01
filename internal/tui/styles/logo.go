package styles

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// RenderLogo returns the ASCII art AISPACE logo rendered with Ayu Dark colors.
func RenderLogo() string {
	logoStyle := lipgloss.NewStyle().Foreground(AccentGreen2)
	subtitleStyle := lipgloss.NewStyle().Foreground(Muted)

	logo := ` █████╗ ██╗███████╗██████╗  █████╗  ██████╗███████╗
██╔══██╗██║██╔════╝██╔══██╗██╔══██╗██╔════╝██╔════╝
███████║██║███████╗██████╔╝███████║██║     █████╗
██╔══██║██║╚════██║██╔═══╝ ██╔══██║██║     ██╔══╝
██║  ██║██║███████║██║     ██║  ██║╚██████╗███████╗
╚═╝  ╚═╝╚═╝╚══════╝╚═╝     ╚═╝  ╚═╝ ╚═════╝╚══════╝`

	return logoStyle.Render(logo) + "\n" + subtitleStyle.Render("by ISMARTZDEV")
}

// Tagline returns the version tagline string.
func Tagline(version string) string {
	return fmt.Sprintf("✦ AI Coding Session Manager ✦  v%s  ✦ Think • Build • Evolve ✦", version)
}
