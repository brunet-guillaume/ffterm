package main

import (
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

type view int

const (
	viewFiles view = iota
	viewPresets
	viewConfirm
)

type model struct {
	view          view
	cursor        int
	entries       []entry
	dir           string
	info          *mediaInfo
	preview       string
	err           string
	message       string
	presetCursor  int
	selectedFile  string
	outputPath    string
	commandArgs   []string
	width         int
	height        int
	scroll        int
	showHidden    bool
	showOnlyMedia bool
	focusedPanel  int
	useNerdFonts  bool
}

func initialModel() model {
	dir, _ := os.Getwd()
	cfg := loadConfig()
	return model{
		view:         viewFiles,
		dir:          dir,
		showHidden:   false,
		entries:      listFiles(dir, false, false),
		useNerdFonts: cfg.UseNerdFonts,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		m.err = ""
		m.message = ""

		switch m.view {
		case viewFiles:
			return m.updateFiles(msg)
		case viewPresets:
			return m.updatePresets(msg)
		case viewConfirm:
			return m.updateConfirm(msg)
		}
	}
	return m, nil
}

func (m *model) updateScroll(listHeight int) {
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	}
	if m.cursor >= m.scroll+listHeight {
		m.scroll = m.cursor - listHeight + 1
	}
}

func (m *model) loadPreview(previewWidth, previewHeight int) {
	if len(m.entries) == 0 {
		return
	}

	selected := m.entries[m.cursor]
	path := filepath.Join(m.dir, selected.name)

	switch selected.mediaType {
	case mediaTypeImage:
		m.preview = "\n" + previewImage(path, previewWidth, previewHeight)
	case mediaTypeVideo:
		m.preview = "\n" + previewVideo(path, previewWidth, previewHeight)
	case mediaTypeAudio:
		m.preview = "\n" + previewAudio(path, previewWidth, previewHeight)
	default:
		m.preview = ""
	}
}

func (m *model) loadCurrentFileInfo(previewWidth, previewHeight int) {
	if len(m.entries) == 0 {
		return
	}

	selected := m.entries[m.cursor]
	if selected.mediaType == mediaTypeNone || selected.isDir {
		m.info = nil
		m.preview = ""
		return
	}

	path := filepath.Join(m.dir, selected.name)

	info, err := probe(path)
	if err != nil {
		m.info = nil
	} else {
		m.info = info
	}

	m.preview = ""
}

func (m model) updateFiles(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	listHeight := m.height - 8
	previewWidth := 35
	previewHeight := 9

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.focusedPanel == 1 {
			if m.presetCursor > 0 {
				m.presetCursor--
			}
		} else {
			if m.cursor > 0 {
				m.cursor--
				m.updateScroll(listHeight)
				m.loadCurrentFileInfo(previewWidth, previewHeight)
			}
		}
	case "down", "j":
		if m.focusedPanel == 1 {
			if m.presetCursor < len(presets)-1 {
				m.presetCursor++
			}
		} else {
			if m.cursor < len(m.entries)-1 {
				m.cursor++
				m.updateScroll(listHeight)
				m.loadCurrentFileInfo(previewWidth, previewHeight)
			}
		}
	case "enter", "l":
		if m.focusedPanel == 1 {
			if len(m.entries) > 0 {
				selected := m.entries[m.cursor]
				if selected.mediaType != mediaTypeNone {
					m.selectedFile = filepath.Join(m.dir, selected.name)
					p := presets[m.presetCursor]
					m.outputPath = buildOutputPath(m.selectedFile, p)
					m.commandArgs = buildCommand(m.selectedFile, m.outputPath, p)
					m.view = viewConfirm
				}
			}
		} else if len(m.entries) > 0 {
			selected := m.entries[m.cursor]
			if selected.isDir {
				if selected.name == ".." {
					m.dir = filepath.Dir(m.dir)
				} else {
					m.dir = filepath.Join(m.dir, selected.name)
				}
				m.entries = listFiles(m.dir, m.showHidden, m.showOnlyMedia)
				m.cursor = 0
				m.scroll = 0
				m.loadCurrentFileInfo(previewWidth, previewHeight)
			}
		}
	case "backspace", "h":
		if m.dir != "/" {
			m.dir = filepath.Dir(m.dir)
			m.entries = listFiles(m.dir, m.showHidden, m.showOnlyMedia)
			m.cursor = 0
			m.scroll = 0
			m.loadCurrentFileInfo(previewWidth, previewHeight)
		}
	case "i":
		if len(m.entries) > 0 {
			selected := m.entries[m.cursor]
			if selected.mediaType != mediaTypeNone {
				path := filepath.Join(m.dir, selected.name)
				info, err := probe(path)
				if err != nil {
					m.err = err.Error()
					m.info = nil
				} else {
					m.info = info
				}
			}
		}
	case "p":
		m.loadPreview(35, 9)
	case "o":
		if len(m.entries) > 0 {
			selected := m.entries[m.cursor]
			if selected.mediaType != mediaTypeNone {
				path := filepath.Join(m.dir, selected.name)
				if err := openFile(path, selected.mediaType); err != nil {
					m.err = err.Error()
				}
			}
		}
	case "c":
		if len(m.entries) > 0 {
			selected := m.entries[m.cursor]
			if selected.mediaType != mediaTypeNone {
				m.selectedFile = filepath.Join(m.dir, selected.name)
				m.view = viewPresets
				m.presetCursor = 0
			}
		}
	case ".":
		m.showHidden = !m.showHidden
		m.entries = listFiles(m.dir, m.showHidden, m.showOnlyMedia)
		m.cursor = 0
		m.scroll = 0
		m.loadCurrentFileInfo(previewWidth, previewHeight)
	case "f":
		m.showOnlyMedia = !m.showOnlyMedia
		m.entries = listFiles(m.dir, m.showHidden, m.showOnlyMedia)
		m.cursor = 0
		m.scroll = 0
		m.loadCurrentFileInfo(previewWidth, previewHeight)
	case "tab":
		m.focusedPanel = (m.focusedPanel + 1) % 3
	case "n":
		m.useNerdFonts = !m.useNerdFonts
		saveConfig(config{UseNerdFonts: m.useNerdFonts})
	}
	return m, nil
}

func (m model) updatePresets(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "escape", "backspace", "h":
		m.view = viewFiles
	case "up", "k":
		if m.presetCursor > 0 {
			m.presetCursor--
		}
	case "down", "j":
		if m.presetCursor < len(presets)-1 {
			m.presetCursor++
		}
	case "enter", "l":
		p := presets[m.presetCursor]
		m.outputPath = buildOutputPath(m.selectedFile, p)
		m.commandArgs = buildCommand(m.selectedFile, m.outputPath, p)
		m.view = viewConfirm
	}
	return m, nil
}

func (m model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "escape", "backspace", "h":
		m.view = viewPresets
	case "enter", "y":
		cmd := exec.Command("ffmpeg", m.commandArgs...)
		err := cmd.Run()
		if err != nil {
			m.err = err.Error()
		} else {
			m.message = "Conversion complete!"
			m.entries = listFiles(m.dir, m.showHidden, m.showOnlyMedia)
		}
		m.view = viewFiles
	case "n":
		m.view = viewFiles
	}
	return m, nil
}

func (m model) View() string {
	switch m.view {
	case viewPresets:
		return m.viewPresets()
	case viewConfirm:
		return m.viewConfirm()
	default:
		return m.viewFiles()
	}
}
