package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func createTopBorder(w int, title string, borderColor lipgloss.Color) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(borderColor).
		Bold(true)

	titleText := titleStyle.Render(" " + title + " ")
	borderChar := lipgloss.NewStyle().Foreground(borderColor).Render("─")
	cornerLeft := lipgloss.NewStyle().Foreground(borderColor).Render("╭")
	cornerRight := lipgloss.NewStyle().Foreground(borderColor).Render("╮")

	titleLen := len(title) + 2
	leftPad := 1
	rightPad := max(0, w-titleLen-leftPad)

	topBorder := cornerLeft +
		strings.Repeat(borderChar, leftPad) +
		titleText +
		strings.Repeat(borderChar, rightPad) +
		cornerRight

	return topBorder
}

func (m model) viewFiles() string {
	var s strings.Builder

	// Calculate dimensions
	filePanelWidth := max(30, m.width/4)
	filePanelHeight := max(5, m.height-3)

	rightTotalWidth := m.width - filePanelWidth - 4
	convertPanelWidth := rightTotalWidth
	convertPanelHeight := max(5, m.height-3-13)

	infoPanelWidth := rightTotalWidth

	// Left panel: files
	fileContentStr := m.renderFileList(filePanelWidth, filePanelHeight)

	// Info panel (with preview)
	infoContentStr := m.renderInfoPanel(infoPanelWidth)

	// Converter panel
	convertContentStr := m.renderConverterPanel(convertPanelHeight)

	// Styles based on focus
	var filePanelStyle, infoPanelStyle, convertPanelStyle lipgloss.Style
	var fileBorderColor, infoBorderColor, convertBorderColor lipgloss.Color

	switch m.focusedPanel {
	case 0:
		filePanelStyle = panelStyle
		infoPanelStyle = panelStyleUnfocused
		convertPanelStyle = panelStyleUnfocused
		fileBorderColor = focusedColor
		infoBorderColor = unfocusedColor
		convertBorderColor = unfocusedColor
	case 1:
		filePanelStyle = panelStyleUnfocused
		infoPanelStyle = panelStyleUnfocused
		convertPanelStyle = panelStyle
		fileBorderColor = unfocusedColor
		infoBorderColor = unfocusedColor
		convertBorderColor = focusedColor
	default:
		filePanelStyle = panelStyleUnfocused
		infoPanelStyle = panelStyle
		convertPanelStyle = panelStyleUnfocused
		fileBorderColor = unfocusedColor
		infoBorderColor = focusedColor
		convertBorderColor = unfocusedColor
	}

	// Build panels
	filePanel := filePanelStyle.Width(filePanelWidth).Height(filePanelHeight).Render(fileContentStr)
	filePanelWithBorder := lipgloss.JoinVertical(lipgloss.Left, createTopBorder(filePanelWidth, "Files explorer", fileBorderColor), filePanel)

	convertPanel := convertPanelStyle.Width(convertPanelWidth).Height(convertPanelHeight).Render(convertContentStr)
	convertPanelWithBorder := lipgloss.JoinVertical(lipgloss.Left, createTopBorder(convertPanelWidth, "Converter", convertBorderColor), convertPanel)

	infoPanel := infoPanelStyle.Width(infoPanelWidth).Render(infoContentStr)
	infoPanelWithBorder := lipgloss.JoinVertical(lipgloss.Left, createTopBorder(infoPanelWidth, "Infos", infoBorderColor), infoPanel)

	// Join Converter + Info vertically
	rightPanels := lipgloss.JoinVertical(lipgloss.Left, convertPanelWithBorder, infoPanelWithBorder)
	// Join Files + Right
	panels := lipgloss.JoinHorizontal(lipgloss.Top, filePanelWithBorder, "", rightPanels)
	s.WriteString(panels + "\n")

	// Messages
	if m.err != "" {
		s.WriteString(errorStyle.Render("⚠️  "+m.err) + "\n")
	}
	if m.message != "" {
		s.WriteString(successStyle.Render("✅ "+m.message) + "\n")
	}

	// Help bar
	s.WriteString(helpStyle.Render("j/k: navigate • o: open • p: preview • c: convert • q: quit"))

	return s.String()
}

func (m model) renderFileList(width, height int) string {
	var content strings.Builder

	for i, e := range m.entries {
		if i < m.scroll || i >= m.scroll+height {
			continue
		}

		var icon string
		var style lipgloss.Style

		if e.isDir {
			if e.name == ".." {
				icon = getIcon("parent", m.useNerdFonts)
			} else {
				icon = getIcon("folder", m.useNerdFonts)
			}
			style = folderStyle
		} else {
			switch e.mediaType {
			case mediaTypeVideo:
				icon = getIcon("video", m.useNerdFonts)
				style = videoStyle
			case mediaTypeAudio:
				icon = getIcon("audio", m.useNerdFonts)
				style = audioStyle
			case mediaTypeImage:
				icon = getIcon("image", m.useNerdFonts)
				style = imageStyle
			default:
				icon = getIcon("file", m.useNerdFonts)
				style = normalStyle
			}
		}

		name := e.name
		maxNameLen := width - 8
		if len(name) > maxNameLen {
			name = name[:maxNameLen-3] + "..."
		}

		line := fmt.Sprintf("%s %s", icon, name)

		if i == m.cursor {
			indicator := "> "
			if m.useNerdFonts {
				indicator = " "
			}
			pipe := lipgloss.NewStyle().Foreground(focusedColor).Render(indicator)
			line = pipe + " " + style.Bold(true).Render(line)
		} else {
			line = "  " + style.Render(line)
		}

		content.WriteString(line + "\n")
	}

	return strings.TrimSuffix(content.String(), "\n")
}

func (m model) renderInfoPanel(totalWidth int) string {
	infoWidth := totalWidth - 50

	// Build info content
	var infoContent strings.Builder

	var currentMediaType mediaType
	if len(m.entries) > 0 && m.cursor < len(m.entries) {
		currentMediaType = m.entries[m.cursor].mediaType
	}

	if m.info != nil {
		if m.info.size != "" {
			infoContent.WriteString(infoLabelStyle.Render("Size:"))
			infoContent.WriteString(infoValueStyle.Render(formatSize(m.info.size)) + "\n")
		}

		if currentMediaType == mediaTypeImage {
			m.renderImageInfo(&infoContent)
		} else {
			m.renderMediaInfo(&infoContent, currentMediaType)
		}
	} else {
		infoContent.WriteString(dirStyle.Render("Select a media file"))
	}

	// Build preview content
	var previewContent string
	if m.preview != "" {
		previewContent = m.preview
	} else if currentMediaType == mediaTypeImage || currentMediaType == mediaTypeAudio || currentMediaType == mediaTypeVideo {
		previewContent = dirStyle.Render("\n\n\n\n\nPress p for preview")
	} else {
		previewContent = ""
	}

	// Create columns with styles
	previewWidth := totalWidth - infoWidth
	infoColumn := lipgloss.NewStyle().Width(infoWidth).Height(11).Render(infoContent.String())
	previewColumn := lipgloss.NewStyle().Width(previewWidth).Align(lipgloss.Center).Render(previewContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, infoColumn, previewColumn)
}

func (m model) renderImageInfo(content *strings.Builder) {
	if m.info.width > 0 && m.info.height > 0 {
		content.WriteString(infoLabelStyle.Render("Dimensions:"))
		content.WriteString(infoValueStyle.Render(formatResolution(m.info.width, m.info.height)) + "\n")
	}
	if m.info.pixFmt != "" {
		content.WriteString(infoLabelStyle.Render("Format:"))
		content.WriteString(infoValueStyle.Render(formatPixelFormat(m.info.pixFmt)) + "\n")
	}
	if m.info.size != "" && m.info.width > 0 && m.info.height > 0 {
		content.WriteString(infoLabelStyle.Render("BPP:"))
		content.WriteString(infoValueStyle.Render(formatBPP(m.info.size, m.info.width, m.info.height)) + "\n")
	}
	content.WriteString(infoLabelStyle.Render("Transparency:"))
	if hasAlpha(m.info.pixFmt) {
		content.WriteString(infoValueStyle.Render("Yes") + "\n")
	} else {
		content.WriteString(infoValueStyle.Render("No") + "\n")
	}
}

func (m model) renderMediaInfo(content *strings.Builder, currentMediaType mediaType) {
	if m.info.duration != "" {
		content.WriteString(infoLabelStyle.Render("Duration:"))
		content.WriteString(infoValueStyle.Render(formatDuration(m.info.duration)) + "\n")
	}
	if m.info.bitrate != "" {
		content.WriteString(infoLabelStyle.Render("Bitrate:"))
		content.WriteString(infoValueStyle.Render(formatBitrate(m.info.bitrate)) + "\n")
	}

	if m.info.videoCodec != "" && currentMediaType == mediaTypeVideo {
		content.WriteString(infoLabelStyle.Render("Codec:"))
		content.WriteString(infoValueStyle.Render(m.info.videoCodec) + "\n")
		content.WriteString(infoLabelStyle.Render("Resolution:"))
		content.WriteString(infoValueStyle.Render(formatResolution(m.info.width, m.info.height)) + "\n")
		if m.info.fps != "" {
			content.WriteString(infoLabelStyle.Render("FPS:"))
			content.WriteString(infoValueStyle.Render(formatFPS(m.info.fps)) + "\n")
		}
		if m.info.isHDR {
			content.WriteString(infoLabelStyle.Render("HDR:"))
			content.WriteString(infoValueStyle.Render("Oui") + "\n")
		}
	}

	if m.info.audioCodec != "" {
		content.WriteString(infoLabelStyle.Render("Codec:"))
		content.WriteString(infoValueStyle.Render(m.info.audioCodec) + "\n")
		if m.info.sampleRate != "" {
			content.WriteString(infoLabelStyle.Render("Sample:"))
			content.WriteString(infoValueStyle.Render(m.info.sampleRate+" Hz") + "\n")
		}
		if m.info.channels > 0 {
			content.WriteString(infoLabelStyle.Render("Channels:"))
			content.WriteString(infoValueStyle.Render(formatChannels(m.info.channels)) + "\n")
		}
	}

	if m.info.audioTracks > 1 || m.info.subTracks > 0 {
		content.WriteString("\n" + panelTitleStyle.Render("Tracks") + "\n")
		if m.info.audioTracks > 1 {
			content.WriteString(infoLabelStyle.Render("Audio:"))
			content.WriteString(infoValueStyle.Render(fmt.Sprintf("%d tracks", m.info.audioTracks)) + "\n")
		}
		if m.info.subTracks > 0 {
			content.WriteString(infoLabelStyle.Render("Subtitles:"))
			content.WriteString(infoValueStyle.Render(fmt.Sprintf("%d tracks", m.info.subTracks)) + "\n")
		}
	}
}

func (m model) renderConverterPanel(height int) string {
	var content strings.Builder

	for i, p := range presets {
		if i >= height-2 {
			break
		}
		if i == m.presetCursor {
			indicator := "> "
			if m.useNerdFonts {
				indicator = " "
			}
			pipe := lipgloss.NewStyle().Foreground(focusedColor).Render(indicator)
			boldName := lipgloss.NewStyle().Bold(true).Render(p.name)
			content.WriteString(pipe + " " + boldName + " " + presetDescStyle.Render(p.description) + "\n")
		} else {
			content.WriteString("  " + p.name + " " + presetDescStyle.Render(p.description) + "\n")
		}
	}

	return content.String()
}

func (m model) viewPresets() string {
	var s strings.Builder

	header := titleStyle.Render("🎬 ffterm") + " " + dirStyle.Render("Convert")
	s.WriteString(header + "\n\n")

	s.WriteString(infoLabelStyle.Render("File:"))
	s.WriteString(infoValueStyle.Render(filepath.Base(m.selectedFile)) + "\n\n")

	panelWidth := m.width - 4
	listHeight := m.height - 10

	var content strings.Builder
	for i, p := range presets {
		if i >= listHeight {
			break
		}

		if i == m.presetCursor {
			indicator := "> "
			if m.useNerdFonts {
				indicator = " "
			}
			pipe := lipgloss.NewStyle().Foreground(focusedColor).Render(indicator)
			boldName := lipgloss.NewStyle().Bold(true).Render(p.name)
			content.WriteString(pipe + boldName + "  " + presetDescStyle.Render(p.description) + "\n")
		} else {
			content.WriteString("  " + p.name + "  " + presetDescStyle.Render(p.description) + "\n")
		}
	}

	panel := panelStyle.Width(panelWidth).Render(content.String())
	s.WriteString(panel + "\n")

	s.WriteString(helpStyle.Render("j/k: navigate • enter: select • backspace: back • q: quit"))

	return s.String()
}

func (m model) viewConfirm() string {
	var s strings.Builder

	header := titleStyle.Render("🎬 ffterm") + " " + dirStyle.Render("Confirm")
	s.WriteString(header + "\n\n")

	panelWidth := m.width - 4

	var content strings.Builder
	content.WriteString(infoLabelStyle.Render("Input:"))
	content.WriteString(infoValueStyle.Render(m.selectedFile) + "\n\n")
	content.WriteString(infoLabelStyle.Render("Output:"))
	content.WriteString(successStyle.Render(m.outputPath) + "\n\n")
	content.WriteString(dirStyle.Render("Command:") + "\n")
	content.WriteString(commandStyle.Render("ffmpeg "+strings.Join(m.commandArgs, " ")) + "\n")

	panel := panelStyle.Width(panelWidth).Render(content.String())
	s.WriteString(panel + "\n")

	s.WriteString(helpStyle.Render("enter/y: execute • n/backspace: cancel • q: quit"))

	return s.String()
}
