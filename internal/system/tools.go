package system

import (
	"context"
	"os/exec"
)

// ToolStatus records whether a CLI tool was found on PATH.
type ToolStatus struct {
	Name      string
	Installed bool
	Path      string
}

// DetectTools checks whether each named tool is available on PATH.
// It performs no subprocess execution — only LookPath calls.
func DetectTools(ctx context.Context, names []string) map[string]ToolStatus {
	result := make(map[string]ToolStatus, len(names))
	for _, name := range names {
		path, err := exec.LookPath(name)
		result[name] = ToolStatus{
			Name:      name,
			Installed: err == nil,
			Path:      path,
		}
	}
	return result
}
