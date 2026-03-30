package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const ManifestFilename = "manifest.json"

// BackupSource identifies what operation created a backup.
type BackupSource string

const (
	BackupSourceInstall BackupSource = "install"
	BackupSourceSync    BackupSource = "sync"
)

// Label returns a human-readable string for the BackupSource.
func (s BackupSource) Label() string {
	switch s {
	case BackupSourceInstall:
		return "install"
	case BackupSourceSync:
		return "sync"
	default:
		return "unknown source"
	}
}

// Manifest is the persisted record of a snapshot set.
type Manifest struct {
	ID          string          `json:"id"`
	CreatedAt   time.Time       `json:"created_at"`
	RootDir     string          `json:"root_dir"`
	Entries     []ManifestEntry `json:"entries"`
	Source      BackupSource    `json:"source,omitempty"`
	Description string          `json:"description,omitempty"`
	FileCount   int             `json:"file_count,omitempty"`
	CreatedByVersion string     `json:"created_by_version,omitempty"`
}

// DisplayLabel returns a human-readable label suitable for display in CLI/TUI.
func (m Manifest) DisplayLabel() string {
	base := fmt.Sprintf("%s — %s", m.Source.Label(), m.CreatedAt.Local().Format("2006-01-02 15:04"))
	if m.FileCount > 0 {
		return fmt.Sprintf("%s (%d files)", base, m.FileCount)
	}
	return base
}

// ManifestEntry records the original and snapshot path of a single file.
type ManifestEntry struct {
	OriginalPath string `json:"original_path"`
	SnapshotPath string `json:"snapshot_path"`
	Existed      bool   `json:"existed"`
	Mode         uint32 `json:"mode,omitempty"`
}

// WriteManifest serializes manifest to path, creating parent dirs as needed.
func WriteManifest(path string, manifest Manifest) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create manifest directory %q: %w", path, err)
	}
	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	content = append(content, '\n')
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write manifest %q: %w", path, err)
	}
	return nil
}

// ReadManifest reads and deserializes a manifest from path.
func ReadManifest(path string) (Manifest, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("read manifest %q: %w", path, err)
	}
	var manifest Manifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("unmarshal manifest %q: %w", path, err)
	}
	return manifest, nil
}

// DeleteBackup removes the entire backup directory.
func DeleteBackup(manifest Manifest) error {
	if manifest.RootDir == "" {
		return fmt.Errorf("backup has no root directory")
	}
	return os.RemoveAll(manifest.RootDir)
}
