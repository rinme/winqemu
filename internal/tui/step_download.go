package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"winqemu/pkg/downloader"
	"winqemu/pkg/msdl"
)

func (m *AppModel) updateISOSource(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.UseLocalISO {
		var cmd tea.Cmd
		m.LocalISOInput, cmd = m.LocalISOInput.Update(msg)

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				path := strings.TrimSpace(m.LocalISOInput.Value())
				if path == "" {
					m.Err = fmt.Errorf("please provide a path to the ISO")
					return m, nil
				}
				if _, err := os.Stat(path); err != nil {
					m.Err = fmt.Errorf("file not found: %s", path)
					return m, nil
				}
				m.DownloadedISO = path
				m.Err = nil
				m.Step = StepHardware
				m.Cursor = 0
				m.VMNameInput.Focus()
				return m, nil
			case "esc":
				m.UseLocalISO = false
				m.LocalISOInput.Blur()
				return m, nil
			}
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < 1 {
				m.Cursor++
			}
		case "enter":
			if m.Cursor == 0 {
				// Start download
				m.Step = StepDownloadProgress
				m.IsLoading = true
				m.Err = nil
				return m, m.startDownloadCmd()
			} else {
				// Local ISO path
				m.UseLocalISO = true
				m.LocalISOInput.Focus()
			}
		case "b", "left", "h":
			m.Step = StepEditionSelect
			return m, nil
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *AppModel) viewISOSource() string {
	if m.UseLocalISO {
		var rows []string
		rows = append(rows, "Enter path to your local Windows ISO:")
		rows = append(rows, strings.Repeat("─", 65))
		rows = append(rows, m.LocalISOInput.View())

		if m.Err != nil {
			rows = append(rows, ErrorStyle.Render(fmt.Sprintf("\nError: %v", m.Err)))
		}

		card := InfoCard.Render(strings.Join(rows, "\n"))
		help := HelpStyle.Render("Press [Enter] to confirm path • [Esc] to return to download options")
		return lipgloss.JoinVertical(lipgloss.Left, card, help)
	}

	var rows []string
	rows = append(rows, fmt.Sprintf("Installation Media for %s:", m.SelectedEdition.Name))
	rows = append(rows, strings.Repeat("─", 65))

	opts := []string{
		"📥 Download Genuine ISO from Microsoft/Massgrave CDN (Resumable + Verified)",
		"📁 Use an existing local Windows ISO on this machine",
	}

	for i, opt := range opts {
		cursorStr := "  "
		itemStyle := ListItemNormal
		if i == m.Cursor {
			cursorStr = "❯ "
			itemStyle = ListItemSelected
		}
		rows = append(rows, fmt.Sprintf("%s%s", cursorStr, itemStyle.Render(opt)))
	}

	rows = append(rows, "")
	if m.SelectedEdition.MSDLProductID != "" || m.SelectedEdition.MSDLEvalSlug != "" {
		rows = append(rows, "CDN Source:  ⚡ Official Microsoft CDN via MSDL (msdl.tech-latest.com)")
	} else {
		rows = append(rows, fmt.Sprintf("Source Link: %s", m.SelectedEdition.DownloadURL))
	}
	rows = append(rows, fmt.Sprintf("File Size:   ~%.2f GB", float64(m.SelectedEdition.SizeBytes)/(1024*1024*1024)))

	card := InfoCard.Render(strings.Join(rows, "\n"))
	help := HelpStyle.Render("Use [↑/↓] to choose source • [Enter] select • [b] back • [q] quit")

	return lipgloss.JoinVertical(lipgloss.Left, card, help)
}

func (m *AppModel) startDownloadCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		m.cancelDownload = cancel

		home, _ := os.UserHomeDir()
		cacheDir := filepath.Join(home, ".cache", "winqemu", "iso")
		_ = os.MkdirAll(cacheDir, 0755)

		destFile := filepath.Join(cacheDir, fmt.Sprintf("%s-%s.iso", m.SelectedEdition.ID, m.SelectedEdition.Architecture))
		m.DownloadedISO = destFile

		dl := downloader.NewDownloader()
		dl.MinBytes = 100 * 1024 * 1024

		// If already genuinely downloaded (>100MB), skip
		if cached, path, _ := downloader.IsISOCached(m.SelectedEdition.ID, m.SelectedEdition.Architecture, cacheDir); cached {
			m.DownloadedISO = path
			return DownloadDoneMsg{Err: nil}
		}

		activeURL, _ := msdl.ResolveLiveDownloadURL(ctx, m.SelectedEdition)
		if activeURL == "" {
			return DownloadDoneMsg{Err: fmt.Errorf("no download link configured or resolved for edition: %s", m.SelectedEdition.Name)}
		}

		err := dl.Download(ctx, activeURL, destFile, "", func(p downloader.Progress) {
			if m.program != nil {
				m.program.Send(DownloadMsg(p))
			}
		})

		if err != nil {
			return DownloadDoneMsg{Err: err}
		}

		return DownloadDoneMsg{Err: nil}
	}
}

func (m *AppModel) updateDownloadProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "esc" {
			if m.cancelDownload != nil {
				m.cancelDownload()
			}
			m.Step = StepISOSource
			m.IsLoading = false
			return m, nil
		}
	}
	return m, nil
}

func (m *AppModel) viewDownloadProgress() string {
	var rows []string
	rows = append(rows, "Downloading Genuine Windows Installation Media...")
	rows = append(rows, strings.Repeat("─", 65))
	rows = append(rows, fmt.Sprintf("Edition:     %s", m.SelectedEdition.Name))
	rows = append(rows, fmt.Sprintf("Destination: %s", m.DownloadedISO))
	rows = append(rows, "")
	rows = append(rows, m.ProgressBar.View())
	rows = append(rows, "")
	if m.StatusMsg != "" {
		rows = append(rows, m.StatusMsg)
	} else {
		rows = append(rows, "Connecting to CDN endpoint...")
	}

	card := InfoCard.Render(strings.Join(rows, "\n"))
	help := HelpStyle.Render("Press [Esc] to cancel download")

	return lipgloss.JoinVertical(lipgloss.Left, card, help)
}
