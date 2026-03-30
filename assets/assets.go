package assets

import "embed"

//go:embed all:personas all:claude all:opencode all:skills
var FS embed.FS

// Read returns the content of an embedded asset file.
// path is relative to the assets/ directory (e.g. "personas/neutral.md").
func Read(path string) (string, error) {
	data, err := FS.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MustRead returns the content of an embedded asset file or panics.
func MustRead(path string) string {
	s, err := Read(path)
	if err != nil {
		panic("assets: " + err.Error())
	}
	return s
}
