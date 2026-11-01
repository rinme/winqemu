package vm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ExistingVM represents an existing virtual machine discovered on the host.
type ExistingVM struct {
	Name             string    `json:"name"`
	Path             string    `json:"path"`
	DiskPath         string    `json:"disk_path"`
	DiskSizeMB       int64     `json:"disk_size_mb"`
	LaunchScriptPath string    `json:"launch_script_path"`
	LibvirtXMLPath   string    `json:"libvirt_xml_path"`
	IsLibvirtDefined bool      `json:"is_libvirt_defined"`
	CreatedTime      time.Time `json:"created_time"`
}

// GetDefaultStorageDir returns the default storage root directory for WinQEMU VMs.
func GetDefaultStorageDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/var/tmp/winqemu/vms"
	}
	return filepath.Join(home, ".local", "share", "winqemu", "vms")
}

// ScanExistingVMs examines a target directory (or default if empty) and discovers configured VMs.
func ScanExistingVMs(storageDir string) ([]ExistingVM, error) {
	if storageDir == "" {
		storageDir = GetDefaultStorageDir()
	}

	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		return []ExistingVM{}, nil
	}

	entries, err := os.ReadDir(storageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage directory: %w", err)
	}

	var vms []ExistingVM
	definedDomains := getLibvirtDefinedDomains()

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		vmName := entry.Name()
		vmDir := filepath.Join(storageDir, vmName)

		// Look for disk or launch script
		diskPath := filepath.Join(vmDir, fmt.Sprintf("%s.qcow2", vmName))
		var diskSizeMB int64
		if fi, err := os.Stat(diskPath); err == nil {
			diskSizeMB = fi.Size() / (1024 * 1024)
		} else {
			// Find any .qcow2 file in the directory
			files, _ := os.ReadDir(vmDir)
			for _, f := range files {
				if strings.HasSuffix(f.Name(), ".qcow2") {
					diskPath = filepath.Join(vmDir, f.Name())
					if fi, err := f.Info(); err == nil {
						diskSizeMB = fi.Size() / (1024 * 1024)
					}
					break
				}
			}
		}

		scriptPath := filepath.Join(vmDir, "start-vm.sh")
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			scriptPath = filepath.Join(vmDir, "start-vm.bat")
		}

		xmlPath := filepath.Join(vmDir, fmt.Sprintf("%s.xml", vmName))

		info, _ := entry.Info()
		created := time.Now()
		if info != nil {
			created = info.ModTime()
		}

		isLibvirt := false
		if _, ok := definedDomains[vmName]; ok {
			isLibvirt = true
		}

		vms = append(vms, ExistingVM{
			Name:             vmName,
			Path:             vmDir,
			DiskPath:         diskPath,
			DiskSizeMB:       diskSizeMB,
			LaunchScriptPath: scriptPath,
			LibvirtXMLPath:   xmlPath,
			IsLibvirtDefined: isLibvirt,
			CreatedTime:      created,
		})
	}

	return vms, nil
}

func getLibvirtDefinedDomains() map[string]struct{} {
	domains := make(map[string]struct{})
	cmd := exec.Command("virsh", "list", "--all", "--name")
	out, err := cmd.CombinedOutput()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			name := strings.TrimSpace(line)
			if name != "" {
				domains[name] = struct{}{}
			}
		}
	}
	return domains
}

// LaunchExistingVM starts the VM either via virsh/virt-manager or via its launch script.
func LaunchExistingVM(vm ExistingVM) error {
	if vm.IsLibvirtDefined {
		_ = StartLibvirtDomain(vm.Name)
		return LaunchVirtManager(vm.Name)
	}

	if vm.LaunchScriptPath != "" {
		if _, err := os.Stat(vm.LaunchScriptPath); err == nil {
			cmd := exec.Command("bash", vm.LaunchScriptPath)
			return cmd.Start()
		}
	}

	return fmt.Errorf("no launch script or libvirt domain found for %s", vm.Name)
}
