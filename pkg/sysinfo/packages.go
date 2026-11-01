package sysinfo

import (
	"fmt"
	"os/exec"
)

// GetInstallCommand returns the recommended shell command to install all virtualization tools on the host.
func GetInstallCommand(info *HostInfo) string {
	switch info.OS {
	case OSDarwin:
		return "brew install qemu"
	case OSWindows:
		return "winget install QEMU.QEMU"
	case OSLinux:
		switch info.Distro {
		case DistroArch:
			return "sudo pacman -S --needed qemu-desktop libvirt virt-manager edk2-ovmf swtpm iptables-nft dnsmasq && sudo systemctl enable --now libvirtd && sudo usermod -aG libvirt,kvm $USER"
		case DistroUbuntu, DistroDebian:
			return "sudo apt update && sudo apt install -y qemu-system-x86 qemu-utils libvirt-daemon-system libvirt-clients virt-manager ovmf swtpm && sudo systemctl enable --now libvirtd && sudo usermod -aG libvirt,kvm $USER"
		case DistroFedora:
			return "sudo dnf install -y @virtualization virt-manager swtpm edk2-ovmf && sudo systemctl enable --now libvirtd && sudo usermod -aG libvirt,kvm $USER"
		case DistroOpenSUSE:
			return "sudo zypper install -y patterns-server-kvm_server virt-manager swtpm edk2-ovmf && sudo systemctl enable --now libvirtd && sudo usermod -aG libvirt,kvm $USER"
		case DistroVoid:
			return "sudo xbps-install -Sy qemu libvirt virt-manager && sudo ln -s /etc/sv/libvirtd /var/service/"
		case DistroAlpine:
			return "sudo apk add qemu-system-x86_64 qemu-img libvirt virt-install ovmf swtpm && sudo rc-update add libvirtd"
		default:
			return "sudo apt install -y qemu-system-x86 libvirt-daemon-system virt-manager || sudo pacman -S qemu libvirt virt-manager || sudo dnf install -y @virtualization virt-manager"
		}
	default:
		return "echo 'Please install QEMU and virtualization tools for your operating system.'"
	}
}

// RunInstallCommand executes the installation command interactively using a shell.
func RunInstallCommand(info *HostInfo) error {
	cmdStr := GetInstallCommand(info)
	var cmd *exec.Cmd

	if info.OS == OSWindows {
		cmd = exec.Command("powershell", "-Command", cmdStr)
	} else {
		cmd = exec.Command("bash", "-c", cmdStr)
	}

	cmd.Stdin = nil
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("installation failed: %w (output: %s)", err, string(out))
	}
	return nil
}
