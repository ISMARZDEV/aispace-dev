package skills

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/ismartz/aispace-setup/assets"
	"github.com/ismartz/aispace-setup/internal/agents"
	"github.com/ismartz/aispace-setup/internal/filemerge"
	"github.com/ismartz/aispace-setup/internal/model"
)

type InjectionResult struct {
	Changed bool
	Files   []string
	Skipped []model.SkillID
}

// isSDDSkill reports whether a skill ID belongs to the SDD suite.
// SDD skills are installed by the SDD component to avoid duplicate writes.
func isSDDSkill(id model.SkillID) bool {
	return strings.HasPrefix(string(id), "sdd-")
}

// Inject writes SKILL.md files for each requested skill to the agent's skills directory.
// SDD skills (those beginning with "sdd-") are skipped — the SDD component handles them.
// Missing embedded assets are logged and skipped rather than causing a fatal error.
func Inject(homeDir string, adapter agents.Adapter, skillIDs []model.SkillID) (InjectionResult, error) {
	if !adapter.SupportsSkills() {
		return InjectionResult{Skipped: skillIDs}, nil
	}

	skillDir := adapter.SkillsDir(homeDir)
	if skillDir == "" {
		return InjectionResult{Skipped: skillIDs}, nil
	}

	files := make([]string, 0, len(skillIDs))
	skipped := make([]model.SkillID, 0)
	changed := false

	for _, id := range skillIDs {
		if isSDDSkill(id) {
			// SDD component handles these — skip silently.
			continue
		}

		assetPath := "skills/" + string(id) + "/SKILL.md"
		content, err := assets.Read(assetPath)
		if err != nil {
			log.Printf("skills: skipping %q — embedded asset not found: %v", id, err)
			skipped = append(skipped, id)
			continue
		}
		if len(content) == 0 {
			return InjectionResult{}, fmt.Errorf("skill %q: embedded asset is empty — build may be corrupt", id)
		}

		path := filepath.Join(skillDir, string(id), "SKILL.md")
		result, writeErr := filemerge.WriteFileAtomic(path, []byte(content), 0o644)
		if writeErr != nil {
			return InjectionResult{}, fmt.Errorf("skill %q: write failed: %w", id, writeErr)
		}

		changed = changed || result.Changed
		files = append(files, path)
	}

	return InjectionResult{Changed: changed, Files: files, Skipped: skipped}, nil
}

// SkillPathForAgent returns the filesystem path where a skill file would be written.
func SkillPathForAgent(homeDir string, adapter agents.Adapter, id model.SkillID) string {
	skillDir := adapter.SkillsDir(homeDir)
	if skillDir == "" {
		return ""
	}
	return filepath.Join(skillDir, string(id), "SKILL.md")
}
