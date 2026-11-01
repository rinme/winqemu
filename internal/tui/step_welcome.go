package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"winqemu/pkg/vm"
)

func (m *AppModel) updateWelcome(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "right", "l", " ":
			m.Step = StepDepCheck
			m.Cursor = 0
			return m, nil
		case "v", "V":
			m.ExistingVMs, _ = vm.ScanExistingVMs("")
			m.Step = StepVMManager
			m.Cursor = 0
			return m, nil
		case "q", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *AppModel) viewWelcome() string {
	kvmStatus := "❌ Not available"
	if m.Host.KVMSupported {
		kvmStatus = "✅ Available (/dev/kvm enabled)"
	} else if m.Host.HVFSupported {
		kvmStatus = "✅ Apple Hypervisor Framework (HVF)"
	} else if m.Host.WHPXSupported {
		kvmStatus = "✅ Windows Hypervisor Platform (WHPX)"
	}

	details := fmt.Sprintf(
		"Detected Host OS:   %s\n"+
			"Distribution:       %s\n"+
			"Architecture:       %s\n"+
			"Logical Cores:      %d CPUs\n"+
			"Total Host RAM:     %d MB (~%.1f GB)\n"+
			"Hardware Accel:     %s\n",
		m.Host.OS,
		m.Host.DistroPretty,
		m.Host.Arch,
		m.Host.LogicalCPUs,
		m.Host.TotalRAMMB,
		float64(m.Host.TotalRAMMB)/1024.0,
		kvmStatus,
	)

	if len(m.ExistingVMs) > 0 {
		details += fmt.Sprintf("\nExisting VMs:       %s (Press [v] to manage)\n",
			SuccessStyle.Render(fmt.Sprintf("%d virtual machines detected", len(m.ExistingVMs))))
	}

	card := InfoCard.Render(details)
	help := HelpStyle.Render("Press [Enter] to start new VM setup • [v] manage existing VMs • [q] exit")

	return lipgloss.JoinVertical(lipgloss.Left, card, help)
}

func (m *AppModel) updateVMManager(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.ExistingVMs)-1 {
				m.Cursor++
			}
		case "enter", "space":
			if len(m.ExistingVMs) > 0 && m.Cursor < len(m.ExistingVMs) {
				_ = vm.LaunchExistingVM(m.ExistingVMs[m.Cursor])
				return m, tea.Quit
			}
		case "b", "esc":
			m.Step = StepWelcome
			m.Cursor = 0
			return m, nil
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *AppModel) viewVMManager() string {
	var rows []string
	rows = append(rows, "Discovered WinQEMU Virtual Machines:")
	rows = append(rows, strings.Repeat("─", 65))

	if len(m.ExistingVMs) == 0 {
		rows = append(rows, "No virtual machines found in storage directory.")
	} else {
		for i, v := range m.ExistingVMs {
			cursor := "  "
			itemStyle := ListItemNormal
			if i == m.Cursor {
				cursor = "❯ "
				itemStyle = ListItemSelected
			}

			status := "Script only"
			if v.IsLibvirtDefined {
				status = "Libvirt (virt-manager)"
			}

			rows = append(rows, fmt.Sprintf("%s%s (%d MB, %s)", cursor, itemStyle.Render(v.Name), v.DiskSizeMB, status))
		}
	}

	card := InfoCard.Render(strings.Join(rows, "\n"))
	help := HelpStyle.Render("Use [↑/↓] select • [Enter] launch VM • [b] back to Welcome • [q] quit")

	return lipgloss.JoinVertical(lipgloss.Left, card, help)
}
