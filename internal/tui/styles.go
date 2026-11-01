package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Brand Colors
	PrimaryColor   = lipgloss.Color("#7D56F4") // Modern Purple
	SecondaryColor = lipgloss.Color("#04B575") // Vibrant Mint Green
	AccentColor    = lipgloss.Color("#FF5F87") // Coral / Accent
	GoldColor      = lipgloss.Color("#FFD700") // Gold for WispyDocs badge
	SubtleColor    = lipgloss.Color("#626262") // Slate gray
	BgDark         = lipgloss.Color("#1A1A24") // Deep navy/gray
	TextColor      = lipgloss.Color("#FFFFFF") // White text

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			MarginBottom(1)

	BadgeRecommended = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#000000")).
				Background(GoldColor).
				Padding(0, 1).
				MarginLeft(1)

	HeaderBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(1, 2).
			MarginBottom(1)

	StepIndicator = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Bold(true)

	ListItemSelected = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(PrimaryColor).
				Bold(true).
				Padding(0, 1)

	ListItemNormal = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D8D8D8")).
			Padding(0, 1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(SubtleColor).
			MarginTop(1)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF3333")).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Bold(true)

	InfoCard = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(SubtleColor).
			Padding(1).
			Width(68)
)
