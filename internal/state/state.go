package state

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/ismartz/aispace-setup/internal/model"
)

const stateDir = ".aispace-setup"
const stateFile = "state.json"

// Path returns the absolute path to the state file.
func Path(homeDir string) string {
	return filepath.Join(homeDir, stateDir, stateFile)
}

// Read reads and unmarshals the state file.
// Returns an error if the file does not exist or cannot be decoded.
func Read(homeDir string) (model.InstallState, error) {
	data, err := os.ReadFile(Path(homeDir))
	if err != nil {
		return model.InstallState{}, err
	}
	var s model.InstallState
	if err := json.Unmarshal(data, &s); err != nil {
		return model.InstallState{}, err
	}
	return s, nil
}

// Write persists the install state to disk.
// Creates the .aispace-setup directory if needed.
func Write(homeDir string, state model.InstallState) error {
	dir := filepath.Join(homeDir, stateDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(homeDir), append(data, '\n'), 0o644)
}

// Exists reports whether a state file exists for the given home directory.
func Exists(homeDir string) bool {
	_, err := os.Stat(Path(homeDir))
	return err == nil
}
