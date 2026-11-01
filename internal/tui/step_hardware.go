package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *AppModel) updateHardware(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			m.Cursor = (m.Cursor + 1) % 4
			m.updateFocus()
			return m, nil
		case "shift+tab", "up":
			m.Cursor = (m.Cursor + 3) % 4
			m.updateFocus()
			return m, nil
		case "enter":
			if m.Cursor < 3 {
				m.Cursor++
				m.updateFocus()
				return m, nil
			} else {
				m.Step = StepTuning
				m.Cursor = 0
				return m, nil
			}
		case "b", "esc":
			m.Step = StepISOSource
			m.Cursor = 0
			return m, nil
		}
	}

	var cmd tea.Cmd
	switch m.Cursor {
	case 0:
		m.VMNameInput, cmd = m.VMNameInput.Update(msg)
	case 1:
		m.RAMInput, cmd = m.RAMInput.Update(msg)
	case 2:
		m.CPUInput, cmd = m.CPUInput.Update(msg)
	case 3:
		m.DiskInput, cmd = m.DiskInput.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *AppModel) updateFocus() {
	m.VMNameInput.Blur()
	m.RAMInput.Blur()
	m.CPUInput.Blur()
	m.DiskInput.Blur()

	switch m.Cursor {
	case 0:
		m.VMNameInput.Focus()
	case 1:
		m.RAMInput.Focus()
	case 2:
		m.CPUInput.Focus()
	case 3:
		m.DiskInput.Focus()
	}
}

func (m *AppModel) viewHardware() string {
	var rows []string
	rows = append(rows, "Hardware Resource Allocation (Auto-Calculated from Host):")
	rows = append(rows, strings.Repeat("─", 65))

	rows = append(rows, fmt.Sprintf("%-24s %s", "VM Name:", m.VMNameInput.View()))
	rows = append(rows, fmt.Sprintf("%-24s %s MB (Host: %d MB)", "RAM Allocation:", m.RAMInput.View(), m.Host.TotalRAMMB))
	rows = append(rows, fmt.Sprintf("%-24s %s Cores (Host: %d CPUs)", "vCPU Cores:", m.CPUInput.View(), m.Host.LogicalCPUs))
	rows = append(rows, fmt.Sprintf("%-24s %s GB (Sparse qcow2 disk)", "Virtual Disk Size:", m.DiskInput.View()))

	card := InfoCard.Render(strings.Join(rows, "\n"))
	help := HelpStyle.Render("Use [Tab] / [Enter] to cycle fields • [b] back • [q] quit")

	return lipgloss.JoinVertical(lipgloss.Left, card, help)
}
