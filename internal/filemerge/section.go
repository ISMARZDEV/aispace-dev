package filemerge

import (
	"fmt"
	"strings"
)

const markerPrefix = "ai-setup"

func openMarker(sectionID string) string {
	return fmt.Sprintf("<!-- %s:%s -->", markerPrefix, sectionID)
}

func closeMarker(sectionID string) string {
	return fmt.Sprintf("<!-- /%s:%s -->", markerPrefix, sectionID)
}

// InjectMarkdownSection inserts, replaces, or removes a named section in a markdown file.
//
// Behavior:
//   - If both markers exist and content is non-empty → replaces content between them.
//   - If both markers exist and content is empty → removes the section and its markers.
//   - If markers don't exist and content is non-empty → appends section at end.
//   - If markers don't exist and content is empty → returns existing unchanged.
func InjectMarkdownSection(existing, sectionID, content string) string {
	open := openMarker(sectionID)
	close := closeMarker(sectionID)

	openIdx := strings.Index(existing, open)
	closeIdx := strings.Index(existing, close)

	markersPresent := openIdx != -1 && closeIdx != -1

	if markersPresent {
		if content == "" {
			// Remove section: cut from open marker to end of close marker line.
			before := strings.TrimRight(existing[:openIdx], "\n")
			after := existing[closeIdx+len(close):]
			after = strings.TrimLeft(after, "\n")
			if before == "" {
				return after
			}
			if after == "" {
				return before
			}
			return before + "\n\n" + after
		}

		// Replace content between markers.
		before := existing[:openIdx]
		after := existing[closeIdx+len(close):]
		return before + open + "\n" + strings.TrimSpace(content) + "\n" + close + after
	}

	if content == "" {
		return existing
	}

	// Append new section at end.
	trimmed := strings.TrimRight(existing, "\n")
	section := "\n\n" + open + "\n" + strings.TrimSpace(content) + "\n" + close
	if trimmed == "" {
		return open + "\n" + strings.TrimSpace(content) + "\n" + close
	}
	return trimmed + section
}

// HasSection reports whether the file already contains the named section markers.
func HasSection(content, sectionID string) bool {
	return strings.Contains(content, openMarker(sectionID)) &&
		strings.Contains(content, closeMarker(sectionID))
}
