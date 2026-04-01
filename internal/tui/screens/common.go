package screens

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ismartz/aispace-setup/internal/tui/styles"
)

// renderOptions renders a navigable list with the cursor item highlighted.
func renderOptions(options []string, cursor int) string {
	var b strings.Builder
	for i, opt := range options {
		if i == cursor {
			b.WriteString(styles.SelectedStyle.Render("  ▸ "+opt) + "\n")
		} else {
			b.WriteString(styles.MutedStyle.Render("    "+opt) + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// renderCheckbox renders a single checkbox item.
func renderCheckbox(label string, checked bool, isCursor bool) string {
	mark := "[ ]"
	if checked {
		mark = "[✓]"
	}
	prefix := "  "
	if isCursor {
		prefix = "▸ "
	}
	line := prefix + mark + " " + label
	if checked {
		return styles.CheckedStyle.Render(line)
	}
	if isCursor {
		return styles.CursorStyle.Render(line)
	}
	return styles.UncheckedStyle.Render(line)
}

// renderRadio renders a single radio item.
func renderRadio(label string, selected bool, isCursor bool) string {
	mark := "○"
	if selected {
		mark = "◉"
	}
	prefix := "  "
	if isCursor {
		prefix = "▸ "
	}
	line := prefix + mark + " " + label
	if selected || isCursor {
		return styles.CursorStyle.Render(line)
	}
	return styles.UncheckedStyle.Render(line)
}

// renderSeparator returns a horizontal separator line.
func renderSeparator() string {
	sep := strings.Repeat("━", 50)
	return lipgloss.NewStyle().Foreground(styles.Muted).Render(sep)
}
