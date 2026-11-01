package sysinfo

// OSType represents the operating system family.
type OSType string

const (
	OSLinux   OSType = "linux"
	OSDarwin  OSType = "darwin"
	OSWindows OSType = "windows"
	OSUnknown OSType = "unknown"
)

// LinuxDistro represents specific Linux distributions.
type LinuxDistro string

const (
	DistroArch     LinuxDistro = "arch"
	DistroDebian   LinuxDistro = "debian"
	DistroUbuntu   LinuxDistro = "ubuntu"
	DistroFedora   LinuxDistro = "fedora"
	DistroOpenSUSE LinuxDistro = "opensuse"
	DistroVoid     LinuxDistro = "void"
	DistroAlpine   LinuxDistro = "alpine"
	DistroUnknown  LinuxDistro = "unknown"
)

// PackageManager represents host package managers.
type PackageManager string

const (
	PkgPacman LinuxDistro = "pacman"
	PkgApt    LinuxDistro = "apt"
	PkgDnf    LinuxDistro = "dnf"
	PkgZypper LinuxDistro = "zypper"
	PkgBrew   LinuxDistro = "brew"
	PkgWinget LinuxDistro = "winget"
	PkgChoco  LinuxDistro = "choco"
)

// DependencyStatus describes the presence and paths of required host binaries.
type DependencyStatus struct {
	Name        string `json:"name"`
	Installed   bool   `json:"installed"`
	Path        string `json:"path"`
	PackageName string `json:"package_name"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// HostInfo contains detected host environment details.
type HostInfo struct {
	OS             OSType             `json:"os"`
	Distro         LinuxDistro        `json:"distro"`
	DistroPretty   string             `json:"distro_pretty"`
	Arch           string             `json:"arch"`
	TotalRAMMB     int                `json:"total_ram_mb"`
	AvailableRAMMB int                `json:"available_ram_mb"`
	LogicalCPUs    int                `json:"logical_cpus"`
	KVMSupported   bool               `json:"kvm_supported"`
	HVFSupported   bool               `json:"hvf_supported"`
	WHPXSupported  bool               `json:"whpx_supported"`
	Dependencies   []DependencyStatus `json:"dependencies"`
	MissingDeps    []string           `json:"missing_deps"`
}
