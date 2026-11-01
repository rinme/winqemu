package sysinfo

import (
	"bufio"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// DetectHost examines the operating system, distro, architecture, resources, and installed dependencies.
func DetectHost() *HostInfo {
	info := &HostInfo{
		OS:          detectOS(),
		Distro:      DistroUnknown,
		Arch:        runtime.GOARCH,
		LogicalCPUs: runtime.NumCPU(),
	}

	if info.OS == OSLinux {
		detectLinuxDistro(info)
		info.KVMSupported = checkKVM()
	} else if info.OS == OSDarwin {
		info.DistroPretty = "macOS"
		info.HVFSupported = true
	} else if info.OS == OSWindows {
		info.DistroPretty = "Windows"
		info.WHPXSupported = true
	}

	detectRAM(info)
	checkDependencies(info)

	return info
}

func detectOS() OSType {
	switch runtime.GOOS {
	case "linux":
		return OSLinux
	case "darwin":
		return OSDarwin
	case "windows":
		return OSWindows
	default:
		return OSUnknown
	}
}

func parseOSReleaseContent(content string) (LinuxDistro, string) {
	var id, idLike, prettyName, name string
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		val := strings.Trim(parts[1], `"'`)

		switch key {
		case "ID":
			id = strings.ToLower(val)
		case "ID_LIKE":
			idLike = strings.ToLower(val)
		case "PRETTY_NAME":
			prettyName = val
		case "NAME":
			name = val
		}
	}

	if prettyName == "" {
		prettyName = name
	}

	distro := DistroUnknown
	switch {
	case id == "arch" || strings.Contains(idLike, "arch"):
		distro = DistroArch
	case id == "ubuntu" || strings.Contains(idLike, "ubuntu"):
		distro = DistroUbuntu
	case id == "debian" || strings.Contains(idLike, "debian"):
		distro = DistroDebian
	case id == "fedora" || strings.Contains(idLike, "fedora") || strings.Contains(idLike, "rhel"):
		distro = DistroFedora
	case strings.Contains(id, "suse") || strings.Contains(idLike, "suse"):
		distro = DistroOpenSUSE
	case id == "void":
		distro = DistroVoid
	case id == "alpine":
		distro = DistroAlpine
	}

	return distro, prettyName
}

func detectLinuxDistro(info *HostInfo) {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		data, err = os.ReadFile("/usr/lib/os-release")
	}

	if err == nil {
		distro, pretty := parseOSReleaseContent(string(data))
		info.Distro = distro
		info.DistroPretty = pretty
	} else {
		info.Distro = DistroUnknown
		info.DistroPretty = "Linux (Generic)"
	}
}

func checkKVM() bool {
	if _, err := os.Stat("/dev/kvm"); err == nil {
		return true
	}
	return false
}

func detectRAM(info *HostInfo) {
	if info.OS == OSLinux {
		data, err := os.ReadFile("/proc/meminfo")
		if err == nil {
			scanner := bufio.NewScanner(strings.NewReader(string(data)))
			for scanner.Scan() {
				line := scanner.Text()
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					if fields[0] == "MemTotal:" {
						var kb int
						for _, ch := range fields[1] {
							if ch >= '0' && ch <= '9' {
								kb = kb*10 + int(ch-'0')
							}
						}
						info.TotalRAMMB = kb / 1024
					} else if fields[0] == "MemAvailable:" {
						var kb int
						for _, ch := range fields[1] {
							if ch >= '0' && ch <= '9' {
								kb = kb*10 + int(ch-'0')
							}
						}
						info.AvailableRAMMB = kb / 1024
					}
				}
			}
		}
	}

	// Fallback reasonable default if /proc/meminfo isn't available
	if info.TotalRAMMB == 0 {
		info.TotalRAMMB = 16384
		info.AvailableRAMMB = 8192
	}
}

func checkDependencies(info *HostInfo) {
	var depList []DependencyStatus

	if info.OS == OSLinux {
		depList = []DependencyStatus{
			{Name: "qemu-system-x86_64", Description: "QEMU Virtual Machine Emulator", Required: true},
			{Name: "qemu-img", Description: "QEMU Disk Image Utility", Required: true},
			{Name: "virsh", Description: "Libvirt CLI Management Tool", Required: true},
			{Name: "virt-install", Description: "Command line VM provisioner", Required: true},
			{Name: "virt-manager", Description: "Libvirt GUI Manager", Required: false},
			{Name: "swtpm", Description: "Software TPM Emulator for Win11", Required: false},
		}
	} else if info.OS == OSDarwin {
		depList = []DependencyStatus{
			{Name: "qemu-system-x86_64", Description: "QEMU Virtual Machine Emulator", Required: true},
			{Name: "qemu-img", Description: "QEMU Disk Image Utility", Required: true},
		}
	} else if info.OS == OSWindows {
		depList = []DependencyStatus{
			{Name: "qemu-system-x86_64.exe", Description: "QEMU Virtual Machine Emulator", Required: true},
			{Name: "qemu-img.exe", Description: "QEMU Disk Image Utility", Required: true},
		}
	}

	for i := range depList {
		path, err := exec.LookPath(depList[i].Name)
		if err == nil {
			depList[i].Installed = true
			depList[i].Path = path
		} else {
			depList[i].Installed = false
			if depList[i].Required {
				info.MissingDeps = append(info.MissingDeps, depList[i].Name)
			}
		}
	}

	info.Dependencies = depList
}
