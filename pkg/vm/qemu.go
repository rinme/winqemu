package vm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// BuildQEMUArgs constructs the command-line flags to start QEMU with all configured hardware and tuning.
func BuildQEMUArgs(cfg *VMConfig) []string {
	var args []string

	// Machine & Acceleration
	switch runtime.GOOS {
	case "darwin":
		args = append(args, "-machine", "q35,accel=hvf")
	case "windows":
		args = append(args, "-machine", "q35,accel=whpx")
	default:
		// Linux default
		args = append(args, "-machine", "q35,accel=kvm")
	}

	// Memory & CPU
	args = append(args, "-m", fmt.Sprintf("%dM", cfg.RAMMB))

	cpuFlag := "host"
	if cfg.Tuning.HostCPUPassthrough {
		var hvFlags []string
		if cfg.Tuning.HyperV.Relaxed {
			hvFlags = append(hvFlags, "hv_relaxed")
		}
		if cfg.Tuning.HyperV.VAPIC {
			hvFlags = append(hvFlags, "hv_vapic")
		}
		if cfg.Tuning.HyperV.Spinlocks {
			hvFlags = append(hvFlags, fmt.Sprintf("hv_spinlocks=0x%x", cfg.Tuning.HyperV.SpinlockRetries))
		}
		if cfg.Tuning.HyperV.Time {
			hvFlags = append(hvFlags, "hv_time")
		}
		if cfg.Tuning.HyperV.Synic {
			hvFlags = append(hvFlags, "hv_synic")
		}
		if cfg.Tuning.HyperV.STimer {
			hvFlags = append(hvFlags, "hv_stimer")
		}
		if cfg.Tuning.HyperV.Reset {
			hvFlags = append(hvFlags, "hv_reset")
		}
		if cfg.Tuning.HyperV.Frequencies {
			hvFlags = append(hvFlags, "hv_frequencies")
		}
		if cfg.Tuning.HyperV.Reenlightenment {
			hvFlags = append(hvFlags, "hv_reenlightenment")
		}

		if len(hvFlags) > 0 {
			cpuFlag = fmt.Sprintf("host,%s", strings.Join(hvFlags, ","))
		}
	}

	args = append(args, "-cpu", cpuFlag)
	args = append(args, "-smp", fmt.Sprintf("cpus=%d,sockets=1,cores=%d,threads=%d",
		cfg.CPUCores*cfg.CPUThreads, cfg.CPUCores, cfg.CPUThreads))

	// Main Storage Disk
	if cfg.Tuning.VirtIOSCSI {
		args = append(args, "-device", "virtio-scsi-pci,id=scsi0")
		args = append(args, "-drive", fmt.Sprintf("file=%s,if=none,id=hd0,format=qcow2,cache=none,aio=native", cfg.DiskPath))
		args = append(args, "-device", "scsi-hd,drive=hd0")
	} else {
		args = append(args, "-drive", fmt.Sprintf("file=%s,format=qcow2,if=ide", cfg.DiskPath))
	}

	// Windows ISO installation media
	if cfg.ISOPath != "" {
		args = append(args, "-drive", fmt.Sprintf("file=%s,media=cdrom,index=1", cfg.ISOPath))
	}

	// VirtIO Drivers ISO
	if cfg.VirtIOISOPath != "" {
		args = append(args, "-drive", fmt.Sprintf("file=%s,media=cdrom,index=2", cfg.VirtIOISOPath))
	}

	// Unattend Floppy / Driver Disk
	if cfg.UnattendDisk != "" {
		args = append(args, "-fda", cfg.UnattendDisk)
	}

	// Network
	if cfg.Tuning.VirtIONetwork {
		args = append(args, "-netdev", "user,id=net0", "-device", "virtio-net-pci,netdev=net0")
	} else {
		args = append(args, "-netdev", "user,id=net0", "-device", "e1000,netdev=net0")
	}

	// Display & Audio
	args = append(args, "-vga", "qxl")
	if cfg.Tuning.SpiceGuestTools {
		args = append(args, "-device", "virtio-serial-pci")
		args = append(args, "-chardev", "spicevmc,id=vdagent,name=vdagent")
		args = append(args, "-device", "virtserialport,chardev=vdagent,name=com.redhat.spice.0")
	}

	args = append(args, "-usb", "-device", "usb-tablet")

	// Looking Glass IVSHMEM shared memory device
	if cfg.Tuning.LookingGlassEnabled {
		memMB := cfg.Tuning.LookingGlassMemMB
		if memMB <= 0 {
			memMB = 64
		}
		args = append(args,
			"-device", "ivshmem-plain,memdev=ivshmem0",
			"-object", fmt.Sprintf("memory-backend-file,id=ivshmem0,share=on,mem-path=/dev/shm/looking-glass,size=%dM", memMB),
		)
	}

	return args
}

// GenerateLaunchScript writes an executable shell or batch script to launch the QEMU VM.
func GenerateLaunchScript(cfg *VMConfig) (string, error) {
	args := BuildQEMUArgs(cfg)

	var scriptContent string
	var scriptPath string

	if runtime.GOOS == "windows" {
		scriptPath = filepath.Join(cfg.OutputDir, "start-vm.bat")
		scriptContent = fmt.Sprintf("@echo off\r\nqemu-system-x86_64.exe %s\r\n", strings.Join(args, " "))
	} else {
		scriptPath = filepath.Join(cfg.OutputDir, "start-vm.sh")
		scriptContent = fmt.Sprintf("#!/usr/bin/env bash\nexec qemu-system-x86_64 \\\n  %s\n", strings.Join(args, " \\\n  "))
	}

	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		return "", fmt.Errorf("failed to write launch script: %w", err)
	}

	return scriptPath, nil
}

// CreateQCOW2 creates a sparse qcow2 virtual disk file using qemu-img.
func CreateQCOW2(path string, sizeGB int) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", path, fmt.Sprintf("%dG", sizeGB))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("qemu-img creation failed: %w (output: %s)", err, string(out))
	}

	return nil
}
