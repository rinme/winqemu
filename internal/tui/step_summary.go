package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"winqemu/pkg/unattend"
	"winqemu/pkg/vm"
)

func (m *AppModel) startCreationCmd() tea.Cmd {
	return func() tea.Msg {
		vmName := strings.TrimSpace(m.VMNameInput.Value())
		if vmName == "" {
			vmName = "Windows-VM"
		}

		ramMB, _ := strconv.Atoi(m.RAMInput.Value())
		if ramMB <= 0 {
			ramMB = 8192
		}

		cpuCores, _ := strconv.Atoi(m.CPUInput.Value())
		if cpuCores <= 0 {
			cpuCores = 4
		}

		diskGB, _ := strconv.Atoi(m.DiskInput.Value())
		if diskGB <= 0 {
			diskGB = 64
		}

		home, _ := os.UserHomeDir()
		vmDir := filepath.Join(home, ".local", "share", "winqemu", "vms", vmName)
		_ = os.MkdirAll(vmDir, 0755)

		// 1. Generate autounattend.xml and floppy image
		xmlContent, err := unattend.GenerateAutounattendXML(*m.SelectedEdition, m.TuningConfig, "User", "")
		if err != nil {
			return VMCreateDoneMsg{Err: fmt.Errorf("failed to generate autounattend: %w", err)}
		}

		_, floppyPath, err := unattend.WriteUnattendFiles(vmDir, xmlContent)
		if err != nil {
			return VMCreateDoneMsg{Err: fmt.Errorf("failed to write unattend floppy: %w", err)}
		}

		// 2. Create qcow2 disk
		diskPath := filepath.Join(vmDir, fmt.Sprintf("%s.qcow2", vmName))
		if _, err := os.Stat(diskPath); os.IsNotExist(err) {
			if err := vm.CreateQCOW2(diskPath, diskGB); err != nil {
				return VMCreateDoneMsg{Err: fmt.Errorf("failed to create disk: %w", err)}
			}
		}

		// 3. Build VM Config
		cfg := &vm.VMConfig{
			Name:            vmName,
			RAMMB:           ramMB,
			CPUCores:        cpuCores,
			CPUThreads:      1,
			DiskSizeGB:      diskGB,
			DiskPath:        diskPath,
			ISOPath:         m.DownloadedISO,
			VirtIOISOPath:   m.VirtIOISO,
			UnattendDisk:    floppyPath,
			OutputDir:       vmDir,
			SelectedEdition: *m.SelectedEdition,
			Tuning:          m.TuningConfig,
			UseLibvirt:      m.Host.OS == "linux",
		}

		// 4. Generate QEMU launch script
		scriptPath, err := vm.GenerateLaunchScript(cfg)
		if err != nil {
			return VMCreateDoneMsg{Err: fmt.Errorf("failed to generate launch script: %w", err)}
		}

		res := &vm.CreationResult{
			VMName:           vmName,
			DiskPath:         diskPath,
			DriverDiskPath:   floppyPath,
			LaunchScriptPath: scriptPath,
		}

		// 5. If Linux, generate Libvirt XML and register with virsh
		if m.Host.OS == "linux" {
			xmlPath, err := vm.WriteLibvirtXML(cfg)
			if err == nil {
				res.LibvirtXMLPath = xmlPath
				if err := vm.RegisterLibvirtDomain(xmlPath); err == nil {
					res.IsLibvirtReady = true
				}
			}
		}

		return VMCreateDoneMsg{Result: res, Err: nil}
	}
}

func (m *AppModel) updateExecution(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handled by top-level Update via VMCreateDoneMsg
	return m, nil
}

func (m *AppModel) viewExecution() string {
	var rows []string
	rows = append(rows, "Provisioning Windows Virtual Machine...")
	rows = append(rows, strings.Repeat("─", 65))
	rows = append(rows, "• Writing automated autounattend.xml answer file")
	rows = append(rows, "• Creating FAT12 virtual floppy driver image")
	rows = append(rows, "• Initializing sparse qcow2 virtual disk")
	rows = append(rows, "• Generating Hyper-V enlightenments & QEMU launcher")
	if m.Host.OS == "linux" {
		rows = append(rows, "• Registering domain with Libvirt (virt-manager integration)")
	}

	card := InfoCard.Render(strings.Join(rows, "\n"))
	return card
}

func (m *AppModel) updateDone(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "l", "L", "enter":
			// Launch VM
			if m.Result != nil {
				if m.Result.IsLibvirtReady {
					_ = vm.LaunchVirtManager(m.Result.VMName)
				} else {
					_ = vm.LaunchVirtManager("")
				}
			}
			return m, tea.Quit
		case "q", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *AppModel) viewDone() string {
	var rows []string
	rows = append(rows, SuccessStyle.Render("🎉 Virtual Machine Successfully Created!"))
	rows = append(rows, strings.Repeat("─", 65))

	if m.Result != nil {
		rows = append(rows, fmt.Sprintf("%-20s %s", "VM Name:", m.Result.VMName))
		rows = append(rows, fmt.Sprintf("%-20s %s", "Virtual Disk:", m.Result.DiskPath))
		rows = append(rows, fmt.Sprintf("%-20s %s", "Driver Disk:", m.Result.DriverDiskPath))
		rows = append(rows, fmt.Sprintf("%-20s %s", "Launch Script:", m.Result.LaunchScriptPath))

		if m.Result.IsLibvirtReady {
			rows = append(rows, fmt.Sprintf("%-20s %s", "Libvirt Domain:", SuccessStyle.Render("Defined & visible in virt-manager")))
			rows = append(rows, fmt.Sprintf("%-20s %s", "Domain XML:", m.Result.LibvirtXMLPath))
		}
	}

	rows = append(rows, "")
	rows = append(rows, "What would you like to do?")
	rows = append(rows, ListItemSelected.Render("▶ Press [L] or [Enter] to launch in virt-manager / QEMU now"))
	rows = append(rows, ListItemNormal.Render("  Press [Q] to exit"))

	card := InfoCard.Render(strings.Join(rows, "\n"))
	return card
}
