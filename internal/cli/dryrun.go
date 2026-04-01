package cli

import (
	"fmt"
	"strings"
)

// RenderDryRun renders a human-readable dry-run preview.
func RenderDryRun(result InstallResult) string {
	var b strings.Builder

	fmt.Fprintln(&b, "ai-setup install — dry-run")
	fmt.Fprintln(&b, "==========================")
	fmt.Fprintf(&b, "Agents:                 %s\n", joinAgentIDs(result.Selection.Agents))
	fmt.Fprintf(&b, "Persona:                %s\n", result.Selection.Persona)
	fmt.Fprintf(&b, "Preset:                 %s\n", result.Selection.Preset)
	if result.Selection.SDDMode != "" {
		fmt.Fprintf(&b, "SDD mode:               %s\n", result.Selection.SDDMode)
	}
	fmt.Fprintf(&b, "Components order:       %s\n", joinComponentIDs(result.Resolved.OrderedComponents))
	if len(result.Resolved.AddedDependencies) > 0 {
		fmt.Fprintf(&b, "Auto-added deps:        %s\n", joinComponentIDs(result.Resolved.AddedDependencies))
	}
	fmt.Fprintf(&b, "Prepare steps:          %d\n", len(result.Plan.Prepare))
	fmt.Fprintf(&b, "Apply steps:            %d\n", len(result.Plan.Apply))

	return strings.TrimRight(b.String(), "\n")
}
