package filemerge

import "encoding/json"

// MergeJSONObjects performs a deep merge of overlay into base.
// overlay values take precedence. Both inputs must be JSON objects.
// Returns the merged JSON bytes or an error.
func MergeJSONObjects(base, overlay []byte) ([]byte, error) {
	var baseMap map[string]any
	if err := json.Unmarshal(base, &baseMap); err != nil {
		return nil, err
	}

	var overlayMap map[string]any
	if err := json.Unmarshal(overlay, &overlayMap); err != nil {
		return nil, err
	}

	merged := deepMerge(baseMap, overlayMap)
	return json.MarshalIndent(merged, "", "  ")
}

func deepMerge(base, overlay map[string]any) map[string]any {
	result := make(map[string]any, len(base))
	for k, v := range base {
		result[k] = v
	}

	for k, overlayVal := range overlay {
		baseVal, exists := result[k]
		if !exists {
			result[k] = overlayVal
			continue
		}

		baseMap, baseIsMap := baseVal.(map[string]any)
		overlayMap, overlayIsMap := overlayVal.(map[string]any)
		if baseIsMap && overlayIsMap {
			result[k] = deepMerge(baseMap, overlayMap)
			continue
		}

		result[k] = overlayVal
	}

	return result
}
