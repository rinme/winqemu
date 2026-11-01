package sysinfo

import (
	"strings"
	"testing"
)

func TestParseOSRelease(t *testing.T) {
	archRelease := `NAME="Arch Linux"
PRETTY_NAME="Arch Linux"
ID=arch
BUILD_ID=rolling
ANSI_COLOR="38;2;23;147;209"
HOME_URL="https://archlinux.org/"
`
	distro, pretty := parseOSReleaseContent(archRelease)
	if distro != DistroArch {
		t.Errorf("expected DistroArch, got %v", distro)
	}
	if !strings.Contains(pretty, "Arch Linux") {
		t.Errorf("expected pretty name to contain Arch Linux, got %s", pretty)
	}

	fedoraRelease := `NAME="Fedora Linux"
VERSION="40 (Workstation Edition)"
ID=fedora
VERSION_ID=40
PRETTY_NAME="Fedora Linux 40 (Workstation Edition)"
`
	distro, pretty = parseOSReleaseContent(fedoraRelease)
	if distro != DistroFedora {
		t.Errorf("expected DistroFedora, got %v", distro)
	}
	if !strings.Contains(pretty, "Fedora Linux") {
		t.Errorf("expected pretty name to contain Fedora, got %s", pretty)
	}

	ubuntuRelease := `NAME="Ubuntu"
VERSION="24.04 LTS (Noble Numbat)"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="Ubuntu 24.04 LTS"
`
	distro, pretty = parseOSReleaseContent(ubuntuRelease)
	if distro != DistroUbuntu {
		t.Errorf("expected DistroUbuntu, got %v", distro)
	}

	suseRelease := `NAME="openSUSE Tumbleweed"
ID="opensuse-tumbleweed"
ID_LIKE="opensuse suse"
PRETTY_NAME="openSUSE Tumbleweed"
`
	distro, pretty = parseOSReleaseContent(suseRelease)
	if distro != DistroOpenSUSE {
		t.Errorf("expected DistroOpenSUSE, got %v", distro)
	}
}

func TestGetInstallCommand(t *testing.T) {
	tests := []struct {
		distro LinuxDistro
		os     OSType
		cmdSub string
	}{
		{distro: DistroArch, os: OSLinux, cmdSub: "pacman"},
		{distro: DistroUbuntu, os: OSLinux, cmdSub: "apt"},
		{distro: DistroFedora, os: OSLinux, cmdSub: "dnf"},
		{distro: DistroOpenSUSE, os: OSLinux, cmdSub: "zypper"},
		{distro: DistroUnknown, os: OSDarwin, cmdSub: "brew"},
		{distro: DistroUnknown, os: OSWindows, cmdSub: "winget"},
	}

	for _, tt := range tests {
		info := &HostInfo{
			OS:     tt.os,
			Distro: tt.distro,
			MissingDeps: []string{"qemu-system-x86_64", "virt-manager"},
		}
		cmd := GetInstallCommand(info)
		if !strings.Contains(cmd, tt.cmdSub) {
			t.Errorf("expected install command for %s to contain %s, got: %s", tt.distro, tt.cmdSub, cmd)
		}
	}
}
