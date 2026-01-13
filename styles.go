package main

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED"))

	dirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7C3AED"))

	folderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6"))

	videoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981"))

	audioStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B"))

	imageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EC4899"))

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(0, 1).
			BorderTop(false)

	panelStyleUnfocused = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#4B5563")).
				Padding(0, 1).
				BorderTop(false)

	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED"))

	infoLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Width(15)

	infoValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	commandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6")).
			Background(lipgloss.Color("#1E293B"))

	presetDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Italic(true)
)

// Panel colors
var (
	focusedColor   = lipgloss.Color("#7C3AED")
	unfocusedColor = lipgloss.Color("#4B5563")
)
