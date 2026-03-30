package system

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// PlatformProfile describes the current platform for install command resolution.
type PlatformProfile struct {
	OS             string
	LinuxDistro    string
	PackageManager string
	// NpmWritable is true when npm global prefix is user-writable (nvm/fnm/volta).
	// When false on Linux, sudo is required for global npm installs.
	NpmWritable bool
	Supported   bool
}

// Linux distro identifiers.
const (
	LinuxDistroUnknown = "unknown"
	LinuxDistroUbuntu  = "ubuntu"
	LinuxDistroDebian  = "debian"
	LinuxDistroArch    = "arch"
	LinuxDistroFedora  = "fedora"
)

// Detect returns a PlatformProfile for the current system.
// It checks runtime.GOOS, the npm prefix, and on Linux reads /etc/os-release.
func Detect(ctx context.Context) (PlatformProfile, error) {
	tools := DetectTools(ctx, []string{"brew", "node", "npm"})

	osReleaseContent := ""
	if runtime.GOOS == "linux" {
		data, _ := os.ReadFile("/etc/os-release")
		osReleaseContent = string(data)
	}

	profile := resolvePlatformProfile(runtime.GOOS, osReleaseContent, tools)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return profile, err
	}

	if runtime.GOOS == "windows" {
		profile.NpmWritable = true
	} else {
		profile.NpmWritable = detectNpmWritable(homeDir)
	}

	return profile, nil
}

// MacProfile returns a PlatformProfile for macOS (used in tests and dry-run).
func MacProfile() PlatformProfile {
	return PlatformProfile{
		OS:             "darwin",
		PackageManager: "brew",
		NpmWritable:    true,
		Supported:      true,
	}
}

func resolvePlatformProfile(goos, linuxOSRelease string, tools map[string]ToolStatus) PlatformProfile {
	profile := PlatformProfile{OS: goos}

	switch goos {
	case "darwin":
		profile.PackageManager = "brew"
		profile.Supported = true
	case "linux":
		distro := detectLinuxDistro(linuxOSRelease)
		profile.LinuxDistro = distro

		// Homebrew on Linux takes precedence over native package managers.
		if brew, ok := tools["brew"]; ok && brew.Installed {
			profile.PackageManager = "brew"
			profile.Supported = true
			return profile
		}

		switch distro {
		case LinuxDistroUbuntu, LinuxDistroDebian:
			profile.PackageManager = "apt"
			profile.Supported = true
		case LinuxDistroArch:
			profile.PackageManager = "pacman"
			profile.Supported = true
		case LinuxDistroFedora:
			profile.PackageManager = "dnf"
			profile.Supported = true
		default:
			profile.PackageManager = ""
			profile.Supported = false
		}
	case "windows":
		profile.PackageManager = "winget"
		profile.Supported = true
	default:
		profile.Supported = false
	}

	return profile
}

func detectNpmWritable(homeDir string) bool {
	out, err := exec.Command("npm", "config", "get", "prefix").Output()
	if err != nil {
		return false
	}
	return strings.HasPrefix(strings.TrimSpace(string(out)), homeDir)
}

func detectLinuxDistro(osRelease string) string {
	if strings.TrimSpace(osRelease) == "" {
		return LinuxDistroUnknown
	}

	fields := map[string]string{}
	for _, line := range strings.Split(osRelease, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToUpper(strings.TrimSpace(parts[0]))
		value := strings.Trim(strings.TrimSpace(parts[1]), `"`)
		fields[key] = strings.ToLower(value)
	}

	id := fields["ID"]
	idLike := fields["ID_LIKE"]

	if isUbuntuLike(id, idLike) {
		if id == LinuxDistroDebian {
			return LinuxDistroDebian
		}
		return LinuxDistroUbuntu
	}
	if isArchLike(id, idLike) {
		return LinuxDistroArch
	}
	if isFedoraLike(id, idLike) {
		return LinuxDistroFedora
	}
	return LinuxDistroUnknown
}

func isUbuntuLike(id, idLike string) bool {
	if id == LinuxDistroUbuntu || id == LinuxDistroDebian {
		return true
	}
	for _, token := range strings.Fields(idLike) {
		if token == LinuxDistroUbuntu || token == LinuxDistroDebian {
			return true
		}
	}
	return false
}

func isArchLike(id, idLike string) bool {
	if id == LinuxDistroArch {
		return true
	}
	for _, token := range strings.Fields(idLike) {
		if token == LinuxDistroArch {
			return true
		}
	}
	return false
}

func isFedoraLike(id, idLike string) bool {
	if id == LinuxDistroFedora || id == "rhel" || id == "centos" || id == "rocky" || id == "almalinux" {
		return true
	}
	for _, token := range strings.Fields(idLike) {
		if token == LinuxDistroFedora || token == "rhel" {
			return true
		}
	}
	return false
}
