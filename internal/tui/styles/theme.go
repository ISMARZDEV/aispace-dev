package styles

import "github.com/charmbracelet/lipgloss"

// Ayu Dark palette
var (
	Background = lipgloss.Color("#0a0e14")
	Foreground = lipgloss.Color("#b3b1ad")
	White      = lipgloss.Color("#ffffff")
	Muted      = lipgloss.Color("#686868")

	AccentBlue   = lipgloss.Color("#59c2ff")
	AccentBlue2  = lipgloss.Color("#53bdfa")
	AccentGreen  = lipgloss.Color("#c2d94c")
	AccentGreen2 = lipgloss.Color("#91b362")
	AccentYellow = lipgloss.Color("#ffb454")
	AccentRed    = lipgloss.Color("#f07178")
	AccentCyan   = lipgloss.Color("#95e6cb")
	AccentGold   = lipgloss.Color("#ffee99")
	Border       = lipgloss.Color("#01060e")
)

// Styles
var (
	TitleStyle = lipgloss.NewStyle().Foreground(AccentBlue).Bold(true)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(Background).Background(AccentBlue).Bold(true)

	SuccessStyle = lipgloss.NewStyle().Foreground(AccentGreen)
	ErrorStyle   = lipgloss.NewStyle().Foreground(AccentRed)
	WarningStyle = lipgloss.NewStyle().Foreground(AccentYellow)
	MutedStyle   = lipgloss.NewStyle().Foreground(Muted)
	SubtextStyle = lipgloss.NewStyle().Foreground(Foreground)
	HeadingStyle = lipgloss.NewStyle().Foreground(White).Bold(true)
	HelpStyle    = lipgloss.NewStyle().Foreground(Muted).Italic(true)

	PersonaBadge = lipgloss.NewStyle().
			Foreground(AccentGold).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(AccentGold).
			Padding(0, 1)

	FrameStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(AccentBlue).
			Padding(1, 2)

	CheckedStyle   = lipgloss.NewStyle().Foreground(AccentBlue).Bold(true)
	UncheckedStyle = lipgloss.NewStyle().Foreground(Muted)
	CursorStyle    = lipgloss.NewStyle().Foreground(AccentBlue).Bold(true)
	AccentStyle = lipgloss.NewStyle().Foreground(AccentCyan)
)
