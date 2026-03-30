package filemerge

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// WriteResult reports the outcome of a WriteFileAtomic call.
type WriteResult struct {
	Changed bool // true if the file was created or its content changed
}

// WriteFileAtomic writes data to path using a write-to-tmp-then-rename strategy.
// Returns WriteResult.Changed=true if the file was created or its content changed.
// If the existing file already contains identical bytes, no write is performed.
func WriteFileAtomic(path string, data []byte, perm os.FileMode) (WriteResult, error) {
	// Check existing content to avoid unnecessary writes.
	existing, err := os.ReadFile(path)
	if err == nil && bytes.Equal(existing, data) {
		return WriteResult{Changed: false}, nil
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return WriteResult{}, fmt.Errorf("create parent dir %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, ".aisetup-tmp-*")
	if err != nil {
		return WriteResult{}, fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return WriteResult{}, fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return WriteResult{}, fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Chmod(tmpPath, perm); err != nil {
		os.Remove(tmpPath)
		return WriteResult{}, fmt.Errorf("chmod temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return WriteResult{}, fmt.Errorf("rename to final path: %w", err)
	}

	return WriteResult{Changed: true}, nil
}
