package skills

import "github.com/ismartz/aispace-setup/internal/model"

// SkillsForPreset returns the skill IDs to install for a given preset.
// SDD skills (sdd-*) are intentionally excluded — they are installed by the SDD component.
func SkillsForPreset(preset model.PresetID) []model.SkillID {
	switch preset {
	case model.PresetFull:
		return []model.SkillID{
			model.SkillJudgmentDay,
			model.SkillBranchPR,
			model.SkillIssueCreation,
		}
	case model.PresetCore:
		return []model.SkillID{
			model.SkillJudgmentDay,
			model.SkillBranchPR,
		}
	case model.PresetMinimal:
		return []model.SkillID{}
	case model.PresetCustom:
		return []model.SkillID{}
	default:
		return []model.SkillID{}
	}
}
