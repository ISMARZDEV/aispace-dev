package filemerge_test

import (
	"testing"

	"github.com/ismartz/aispace-setup/internal/filemerge"
)

func TestInjectMarkdownSection_append(t *testing.T) {
	existing := "# My CLAUDE.md\n\nSome existing content."
	result := filemerge.InjectMarkdownSection(existing, "persona", "You are Jarvis.")

	if !contains(result, "<!-- ai-setup:persona -->") {
		t.Error("missing open marker")
	}
	if !contains(result, "<!-- /ai-setup:persona -->") {
		t.Error("missing close marker")
	}
	if !contains(result, "You are Jarvis.") {
		t.Error("missing injected content")
	}
	if !contains(result, "Some existing content.") {
		t.Error("existing content was removed")
	}
}

func TestInjectMarkdownSection_replace(t *testing.T) {
	existing := "# Header\n\n<!-- ai-setup:persona -->\nOld content\n<!-- /ai-setup:persona -->\n\nFooter"
	result := filemerge.InjectMarkdownSection(existing, "persona", "New content")

	if contains(result, "Old content") {
		t.Error("old content should be replaced")
	}
	if !contains(result, "New content") {
		t.Error("new content should be present")
	}
	if !contains(result, "Footer") {
		t.Error("surrounding content should be preserved")
	}
}

func TestInjectMarkdownSection_remove(t *testing.T) {
	existing := "# Header\n\n<!-- ai-setup:persona -->\nOld content\n<!-- /ai-setup:persona -->\n\nFooter"
	result := filemerge.InjectMarkdownSection(existing, "persona", "")

	if contains(result, "ai-setup:persona") {
		t.Error("markers should be removed when content is empty")
	}
	if contains(result, "Old content") {
		t.Error("section content should be removed")
	}
	if !contains(result, "Footer") {
		t.Error("surrounding content should be preserved")
	}
}

func TestInjectMarkdownSection_empty_existing(t *testing.T) {
	result := filemerge.InjectMarkdownSection("", "persona", "Hello")
	if result != "<!-- ai-setup:persona -->\nHello\n<!-- /ai-setup:persona -->" {
		t.Errorf("unexpected result for empty existing: %q", result)
	}
}

func TestHasSection(t *testing.T) {
	content := "<!-- ai-setup:sdd -->\ncontent\n<!-- /ai-setup:sdd -->"
	if !filemerge.HasSection(content, "sdd") {
		t.Error("should detect existing section")
	}
	if filemerge.HasSection(content, "persona") {
		t.Error("should not detect non-existent section")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
