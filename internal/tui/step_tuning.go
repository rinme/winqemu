package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"winqemu/pkg/tuning"
)

func (m *AppModel) updateTuning(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < 6 {
				m.Cursor++
			}
		case " ":
			// Toggle options
			switch m.Cursor {
			case 0:
				// Switch profile
				if m.TuningConfig.Profile == tuning.ProfileGaming {
					m.TuningConfig = tuning.GetPreset(tuning.ProfileBalanced, m.SelectedEdition.IsWin11)
				} else {
					m.TuningConfig = tuning.GetPreset(tuning.ProfileGaming, m.SelectedEdition.IsWin11)
				}
			case 1:
				m.TuningConfig.BypassTPMAndOOBE = !m.TuningConfig.BypassTPMAndOOBE
			case 2:
				m.TuningConfig.LookingGlassEnabled = !m.TuningConfig.LookingGlassEnabled
			case 3:
				m.TuningConfig.VirtIOSCSI = !m.TuningConfig.VirtIOSCSI
			case 4:
				m.TuningConfig.MASActivationPrompt = !m.TuningConfig.MASActivationPrompt
			case 5:
				// Cycle audio
				switch m.TuningConfig.AudioBackend {
				case "scream":
					m.TuningConfig.AudioBackend = "ich9"
				case "ich9":
					m.TuningConfig.AudioBackend = "hda"
				default:
					m.TuningConfig.AudioBackend = "scream"
				}
			}
		case "enter":
			if m.Cursor == 6 {
				m.Step = StepExecution
				m.IsLoading = true
				return m, m.startCreationCmd()
			} else {
				m.Cursor++
			}
		case "b", "esc":
			m.Step = StepHardware
			m.Cursor = 0
			m.updateFocus()
			return m, nil
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *AppModel) viewTuning() string {
	var rows []string
	rows = append(rows, "Performance & Quality of Life Optimization:")
	rows = append(rows, strings.Repeat("─", 65))

	// Profile selection
	profStr := "⚡ Maximum Gaming & Low-Latency Profile"
	if m.TuningConfig.Profile == tuning.ProfileBalanced {
		profStr = "⚖  Balanced Standard Profile"
	}

	cursor0 := "  "
	if m.Cursor == 0 {
		cursor0 = "❯ "
	}
	rows = append(rows, fmt.Sprintf("%sProfile: %s (Press [Space] to switch)", cursor0, lipgloss.NewStyle().Foreground(PrimaryColor).Bold(true).Render(profStr)))

	// Checkboxes
	checkOption := func(idx int, label string, val bool) string {
		cursor := "  "
		if m.Cursor == idx {
			cursor = "❯ "
		}
		box := "[ ]"
		if val {
			box = "[✔]"
		}
		return fmt.Sprintf("%s%s %s", cursor, lipgloss.NewStyle().Foreground(SecondaryColor).Render(box), label)
	}

	rows = append(rows, "")
	rows = append(rows, checkOption(1, "Bypass Win11 TPM 2.0, Secure Boot, & MS Account OOBE", m.TuningConfig.BypassTPMAndOOBE))
	rows = append(rows, checkOption(2, "Looking Glass IVSHMEM Support (Low-Latency Display)", m.TuningConfig.LookingGlassEnabled))
	rows = append(rows, checkOption(3, "VirtIO SCSI Storage & 10G VirtIO Network Drivers", m.TuningConfig.VirtIOSCSI))
	rows = append(rows, checkOption(4, "Run Microsoft Activation Scripts (MAS) on First Logon", m.TuningConfig.MASActivationPrompt))

	cursor5 := "  "
	if m.Cursor == 5 {
		cursor5 = "❯ "
	}
	rows = append(rows, fmt.Sprintf("%sAudio Backend: %s (Press [Space] to cycle)", cursor5,
		lipgloss.NewStyle().Foreground(PrimaryColor).Bold(true).Render(strings.ToUpper(m.TuningConfig.AudioBackend))))

	rows = append(rows, "")
	cursor6 := "  "
	style6 := ListItemNormal
	if m.Cursor == 6 {
		cursor6 = "❯ "
		style6 = ListItemSelected
	}
	rows = append(rows, fmt.Sprintf("%s%s", cursor6, style6.Render("🚀 Build & Configure Virtual Machine")))

	card := InfoCard.Render(strings.Join(rows, "\n"))
	help := HelpStyle.Render("Use [↑/↓] navigate • [Space] toggle/cycle • [Enter] build • [b] back")

	return lipgloss.JoinVertical(lipgloss.Left, card, help)
}
