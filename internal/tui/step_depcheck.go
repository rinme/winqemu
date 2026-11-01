package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"winqemu/pkg/sysinfo"
)

func (m *AppModel) updateDepCheck(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "right", "l":
			m.Step = StepFamilySelect
			m.Cursor = 0
			return m, nil
		case "b", "left", "h":
			m.Step = StepWelcome
			return m, nil
		case "i", "I":
			// Trigger interactive package installation via terminal passthrough
			m.StatusMsg = "Opening interactive terminal for package installation..."
			m.Err = nil
			execCmd := sysinfo.GetInstallExecCmd(m.Host)
			return m, tea.ExecProcess(execCmd, func(err error) tea.Msg {
				return InstallFinishedMsg{Err: err}
			})
		case "q", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *AppModel) viewDepCheck() string {
	var rows []string
	rows = append(rows, "Virtualization Tools & System Services:")
	rows = append(rows, strings.Repeat("─", 65))

	allInstalled := true
	for _, dep := range m.Host.Dependencies {
		status := SuccessStyle.Render("✔ INSTALLED")
		if !dep.Installed {
			allInstalled = false
			if dep.Required {
				status = ErrorStyle.Render("✖ MISSING (Required)")
			} else {
				status = HelpStyle.Render("⚠ OPTIONAL")
			}
		}
		rows = append(rows, fmt.Sprintf("%-22s %-22s %s", dep.Name, dep.Description, status))
	}

	installCmd := sysinfo.GetInstallCommand(m.Host)
	rows = append(rows, "")
	if !allInstalled {
		rows = append(rows, "Suggested Installation Command:")
		rows = append(rows, lipgloss.NewStyle().Foreground(PrimaryColor).Render(installCmd))
	} else {
		rows = append(rows, SuccessStyle.Render("✔ All virtualization components are properly installed and ready."))
	}

	content := strings.Join(rows, "\n")
	card := InfoCard.Render(content)

	var prompt string
	if !allInstalled {
		prompt = HelpStyle.Render("Press [i] to auto-install missing packages • [Enter] to continue anyway • [b] back • [q] quit")
	} else {
		prompt = HelpStyle.Render("Press [Enter] to choose Windows Edition • [b] back • [q] quit")
	}

	var statusMsg string
	if m.StatusMsg != "" {
		statusMsg = SuccessStyle.Render(fmt.Sprintf("\n%s\n", m.StatusMsg))
	}

	var errMsg string
	if m.Err != nil {
		errMsg = ErrorStyle.Render(fmt.Sprintf("\nError: %v\n", m.Err))
	}

	return lipgloss.JoinVertical(lipgloss.Left, card, statusMsg, errMsg, prompt)
}
