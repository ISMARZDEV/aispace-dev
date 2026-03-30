package engram

import (
	"github.com/ismartz/aispace-setup/internal/installcmd"
	"github.com/ismartz/aispace-setup/internal/model"
	"github.com/ismartz/aispace-setup/internal/system"
)

// InstallCommand returns the platform-specific command sequence to install Engram.
// On macOS with brew: tap + install. On other platforms: returns error (binary download required).
func InstallCommand(profile system.PlatformProfile) ([][]string, error) {
	return installcmd.NewResolver().ResolveComponentInstall(profile, model.ComponentEngram)
}
