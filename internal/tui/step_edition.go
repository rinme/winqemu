package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"winqemu/pkg/downloader"
)

// Family Selection
func (m *AppModel) updateFamilySelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Catalog.Families)-1 {
				m.Cursor++
			}
		case "enter", "right", "l":
			if m.Cursor < len(m.Catalog.Families) {
				m.SelectedFamily = &m.Catalog.Families[m.Cursor]
				m.Step = StepEditionSelect
				m.Cursor = 0
			}
			return m, nil
		case "b", "left", "h":
			m.Step = StepDepCheck
			return m, nil
		case "q", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *AppModel) viewFamilySelect() string {
	var rows []string
	rows = append(rows, "Choose Windows Version Family:")
	rows = append(rows, strings.Repeat("─", 65))

	for i, fam := range m.Catalog.Families {
		cursorStr := "  "
		itemStyle := ListItemNormal
		if i == m.Cursor {
			cursorStr = "❯ "
			itemStyle = ListItemSelected
		}
		rows = append(rows, fmt.Sprintf("%s%s", cursorStr, itemStyle.Render(fam.DisplayName)))
		if i == m.Cursor {
			rows = append(rows, lipgloss.NewStyle().Foreground(SubtleColor).MarginLeft(4).Render(fam.Description))
		}
	}

	card := InfoCard.Render(strings.Join(rows, "\n"))
	help := HelpStyle.Render("Use [↑/↓] or [j/k] to navigate • [Enter] select • [b] back • [q] quit")

	return lipgloss.JoinVertical(lipgloss.Left, card, help)
}

// Edition Selection
func (m *AppModel) updateEditionSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.SelectedFamily == nil || len(m.SelectedFamily.Editions) == 0 {
		m.Step = StepFamilySelect
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.SelectedFamily.Editions)-1 {
				m.Cursor++
			}
		case "enter", "right", "l":
			if m.Cursor < len(m.SelectedFamily.Editions) {
				ed := m.SelectedFamily.Editions[m.Cursor]
				m.SelectedEdition = &ed
				// Update default VM Name
				m.VMNameInput.SetValue(strings.ReplaceAll(ed.Name, " ", "-"))
				m.Step = StepISOSource
				m.Cursor = 0
			}
			return m, nil
		case "b", "left", "h":
			m.Step = StepFamilySelect
			m.Cursor = 0
			return m, nil
		case "q", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *AppModel) viewEditionSelect() string {
	var rows []string
	rows = append(rows, fmt.Sprintf("Editions for %s:", m.SelectedFamily.DisplayName))
	rows = append(rows, strings.Repeat("─", 65))

	var currentRec string
	for i, ed := range m.SelectedFamily.Editions {
		cursorStr := "  "
		itemStyle := ListItemNormal
		if i == m.Cursor {
			cursorStr = "❯ "
			itemStyle = ListItemSelected
			currentRec = ed.Recommendation
		}

		cached, _, _ := downloader.IsISOCached(ed.ID, ed.Architecture, "")
		cachedBadge := ""
		if cached {
			cachedBadge = SuccessStyle.Render(" [Cached - Ready]")
		}

		line := fmt.Sprintf("%s%s%s", cursorStr, itemStyle.Render(ed.Name), cachedBadge)
		if ed.Badge == "LATEST" {
			line += lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF88")).Bold(true).Render(" [⚡ 26H2]")
		}
		if ed.MSDLProductID != "" || ed.MSDLEvalSlug != "" {
			line += lipgloss.NewStyle().Foreground(lipgloss.Color("#00DFFF")).Render(" [MSDL]")
		}
		if ed.IsRecommended {
			line += BadgeRecommended.Render(" ⭐ Wispy")
		}
		rows = append(rows, line)
	}

	if currentRec != "" {
		rows = append(rows, "")
		rows = append(rows, lipgloss.NewStyle().Foreground(GoldColor).Bold(true).Render("Recommendation & Source:"))
		rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color("#E0E0E0")).Render(currentRec))
	}

	card := InfoCard.Render(strings.Join(rows, "\n"))
	help := HelpStyle.Render("Use [↑/↓] to browse editions • [Enter] select • [b] back • [q] quit")

	return lipgloss.JoinVertical(lipgloss.Left, card, help)
}
