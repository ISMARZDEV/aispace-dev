package screens

import (
	"fmt"
	"strings"

	"github.com/ismartz/aispace-setup/internal/pipeline"
	"github.com/ismartz/aispace-setup/internal/tui/styles"
)

// RenderResult renders the final result screen.
func RenderResult(
	result pipeline.ExecutionResult,
	filesChanged int,
	syncErr error,
	isSyncMode bool,
) string {
	var b strings.Builder

	if isSyncMode {
		if syncErr != nil {
			b.WriteString(styles.ErrorStyle.Render("✗ Sync Failed"))
			b.WriteString("\n\n")
			b.WriteString(styles.ErrorStyle.Render("Error: " + syncErr.Error()))
			b.WriteString("\n\n")
			b.WriteString(styles.MutedStyle.Render("Run: ai-setup sync  to try again."))
		} else {
			b.WriteString(styles.SuccessStyle.Render("✓ Sync Complete"))
			b.WriteString("\n\n")
			if filesChanged == 0 {
				b.WriteString(styles.SubtextStyle.Render("All files already up to date."))
			} else {
				b.WriteString(styles.SubtextStyle.Render(fmt.Sprintf("%d file(s) updated.", filesChanged)))
			}
		}
	} else {
		if result.Err != nil {
			b.WriteString(styles.ErrorStyle.Render("✗ Installation Failed"))
			b.WriteString("\n\n")
			b.WriteString(styles.ErrorStyle.Render("Error: " + result.Err.Error()))
			b.WriteString("\n\n")
			b.WriteString(styles.MutedStyle.Render("Your files have been restored from backup."))
			b.WriteString("\n")
			b.WriteString(styles.MutedStyle.Render("Run: ai-setup install  to try again."))
		} else {
			b.WriteString(styles.SuccessStyle.Render("✓ Installation Complete"))
			b.WriteString("\n\n")
			b.WriteString(styles.SubtextStyle.Render("All components installed successfully."))
			b.WriteString("\n\n")
			b.WriteString(styles.HeadingStyle.Render("Next steps:"))
			b.WriteString("\n")
			b.WriteString(styles.SubtextStyle.Render("  • Open a new terminal session"))
			b.WriteString("\n")
			b.WriteString(styles.SubtextStyle.Render("  • Claude Code: type /sdd-init to start a feature"))
			b.WriteString("\n")
			b.WriteString(styles.SubtextStyle.Render("  • OpenCode: type /sdd-init to start a feature"))
			b.WriteString("\n")
			b.WriteString(styles.SubtextStyle.Render("  • Run: ai-setup sync  to update configs later"))
		}
	}

	b.WriteString("\n\n")
	b.WriteString(styles.HelpStyle.Render("enter to exit"))

	return styles.FrameStyle.Render(b.String())
}
