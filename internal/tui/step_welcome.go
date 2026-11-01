package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *AppModel) updateWelcome(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "right", "l", " ":
			m.Step = StepDepCheck
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

	card := InfoCard.Render(details)
	help := HelpStyle.Render("Press [Enter] or [Space] to check virtualization dependencies • [q] to exit")

	return lipgloss.JoinVertical(lipgloss.Left, card, help)
}
