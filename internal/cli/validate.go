package cli

import (
	"fmt"
	"strings"

	"github.com/ismartz/aispace-setup/internal/model"
)

// NormalizeSelection validates and fills defaults for a Selection.
// If Preset is empty, defaults to PresetFull.
// If Persona is empty, defaults to PersonaNeutral.
// If SDDMode is empty, defaults to SDDModeSingle.
// Returns error if any AgentID is unrecognized.
func NormalizeSelection(sel model.Selection) (model.Selection, error) {
	if sel.Preset == "" {
		sel.Preset = model.PresetFull
	}
	if sel.Persona == "" {
		sel.Persona = model.PersonaNeutral
	}
	if sel.SDDMode == "" {
		sel.SDDMode = model.SDDModeSingle
	}

	for _, id := range sel.Agents {
		switch id {
		case model.AgentClaudeCode, model.AgentOpenCode:
			// valid
		default:
			return model.Selection{}, fmt.Errorf("unrecognized agent %q", id)
		}
	}

	if len(sel.Components) == 0 {
		sel.Components = presetComponents(sel.Preset)
	}

	return sel, nil
}

// presetComponents returns the ComponentIDs for a given preset.
func presetComponents(preset model.PresetID) []model.ComponentID {
	switch preset {
	case model.PresetFull:
		return []model.ComponentID{
			model.ComponentEngram,
			model.ComponentSDD,
			model.ComponentSkills,
			model.ComponentContext7,
			model.ComponentPersona,
			model.ComponentPermission,
			model.ComponentTheme,
			model.ComponentAISpace,
		}
	case model.PresetCore:
		return []model.ComponentID{
			model.ComponentEngram,
			model.ComponentSDD,
			model.ComponentSkills,
			model.ComponentContext7,
			model.ComponentAISpace,
		}
	case model.PresetMinimal:
		return []model.ComponentID{model.ComponentEngram}
	default: // PresetCustom or unknown — caller provides components
		return []model.ComponentID{}
	}
}

// asAgentIDs converts []string → []model.AgentID.
func asAgentIDs(ids []string) []model.AgentID {
	result := make([]model.AgentID, 0, len(ids))
	for _, id := range ids {
		result = append(result, model.AgentID(id))
	}
	return result
}

// unique deduplicates agent IDs preserving order.
func unique(ids []model.AgentID) []model.AgentID {
	seen := make(map[model.AgentID]struct{}, len(ids))
	result := make([]model.AgentID, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

// joinAgentIDs renders agent IDs as comma-separated string.
func joinAgentIDs(ids []model.AgentID) string {
	if len(ids) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, string(id))
	}
	return strings.Join(parts, ", ")
}

// joinComponentIDs renders component IDs as comma-separated string.
func joinComponentIDs(ids []model.ComponentID) string {
	if len(ids) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, string(id))
	}
	return strings.Join(parts, ", ")
}
