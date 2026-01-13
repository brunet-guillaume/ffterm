package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Config
type config struct {
	UseNerdFonts bool `json:"use_nerd_fonts"`
}

func getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ffterm", "config.json")
}

func loadConfig() config {
	cfg := config{UseNerdFonts: true} // défaut: activé

	data, err := os.ReadFile(getConfigPath())
	if err != nil {
		return cfg
	}

	json.Unmarshal(data, &cfg)
	return cfg
}

func saveConfig(cfg config) error {
	configPath := getConfigPath()
	configDir := filepath.Dir(configPath)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// Icons
var nerdIcons = map[string]string{
	"folder": " ",
	"video":  " ",
	"audio":  " ",
	"image":  " ",
	"file":   " ",
	"parent": " ",
}

var asciiIcons = map[string]string{
	"folder": "[D] ",
	"video":  "[V] ",
	"audio":  "[A] ",
	"image":  "[I] ",
	"file":   "    ",
	"parent": "[^] ",
}

func getIcon(name string, useNerdFonts bool) string {
	if useNerdFonts {
		return nerdIcons[name]
	}
	return asciiIcons[name]
}

func openFile(path string, mType mediaType) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		if mType == mediaTypeVideo {
			cmd = exec.Command("open", path)
		} else {
			cmd = exec.Command("qlmanage", "-p", path)
		}
	case "linux":
		cmd = exec.Command("xdg-open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	default:
		return fmt.Errorf("OS non supporté: %s", runtime.GOOS)
	}
	return cmd.Start()
}

// Styles
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
			Width(12)

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

type mediaType int

const (
	mediaTypeNone mediaType = iota
	mediaTypeVideo
	mediaTypeAudio
	mediaTypeImage
)

type view int

const (
	viewFiles view = iota
	viewPresets
	viewConfirm
)

type entry struct {
	name      string
	isDir     bool
	mediaType mediaType
}

type preset struct {
	name        string
	description string
	extension   string
	args        []string
}

var presets = []preset{
	{
		name:        "MP4 (H.264)",
		description: "Format universel",
		extension:   ".mp4",
		args:        []string{"-c:v", "libx264", "-preset", "medium", "-crf", "23", "-c:a", "aac", "-b:a", "128k"},
	},
	{
		name:        "MP4 (H.265)",
		description: "Meilleure compression",
		extension:   ".mp4",
		args:        []string{"-c:v", "libx265", "-preset", "medium", "-crf", "28", "-c:a", "aac", "-b:a", "128k"},
	},
	{
		name:        "WebM (VP9)",
		description: "Optimisé web",
		extension:   ".webm",
		args:        []string{"-c:v", "libvpx-vp9", "-crf", "30", "-b:v", "0", "-c:a", "libopus", "-b:a", "128k"},
	},
	{
		name:        "MP3 320kbps",
		description: "Audio HQ",
		extension:   ".mp3",
		args:        []string{"-vn", "-c:a", "libmp3lame", "-b:a", "320k"},
	},
	{
		name:        "AAC 256kbps",
		description: "Audio Apple",
		extension:   ".m4a",
		args:        []string{"-vn", "-c:a", "aac", "-b:a", "256k"},
	},
	{
		name:        "FLAC",
		description: "Audio lossless",
		extension:   ".flac",
		args:        []string{"-vn", "-c:a", "flac"},
	},
	{
		name:        "GIF",
		description: "Animé 480px",
		extension:   ".gif",
		args:        []string{"-vf", "fps=10,scale=480:-1:flags=lanczos", "-loop", "0"},
	},
	{
		name:        "Twitter",
		description: "720p optimisé",
		extension:   ".mp4",
		args:        []string{"-c:v", "libx264", "-preset", "slow", "-crf", "24", "-vf", "scale=-2:720", "-c:a", "aac", "-b:a", "128k"},
	},
	{
		name:        "Compression légère",
		description: "Réduit ~30%",
		extension:   ".mp4",
		args:        []string{"-c:v", "libx264", "-preset", "medium", "-crf", "26", "-c:a", "aac", "-b:a", "128k"},
	},
	{
		name:        "Compression forte",
		description: "Réduit ~60%",
		extension:   ".mp4",
		args:        []string{"-c:v", "libx264", "-preset", "slow", "-crf", "32", "-c:a", "aac", "-b:a", "96k"},
	},
}

var videoExtensions = map[string]bool{
	".mp4": true, ".mkv": true, ".avi": true, ".mov": true, ".webm": true,
	".wmv": true, ".flv": true, ".mpeg": true, ".m4v": true, ".3gp": true,
}

var audioExtensions = map[string]bool{
	".mp3": true, ".wav": true, ".flac": true, ".aac": true, ".ogg": true,
	".m4a": true, ".wma": true, ".opus": true,
}

var imageExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	".bmp": true, ".tiff": true, ".svg": true,
}

func getMediaType(name string) mediaType {
	ext := strings.ToLower(filepath.Ext(name))
	if videoExtensions[ext] {
		return mediaTypeVideo
	}
	if audioExtensions[ext] {
		return mediaTypeAudio
	}
	if imageExtensions[ext] {
		return mediaTypeImage
	}
	return mediaTypeNone
}

type probeResult struct {
	Format struct {
		Duration string `json:"duration"`
		Size     string `json:"size"`
		BitRate  string `json:"bit_rate"`
	} `json:"format"`
	Streams []struct {
		CodecType      string `json:"codec_type"`
		CodecName      string `json:"codec_name"`
		Width          int    `json:"width"`
		Height         int    `json:"height"`
		SampleRate     string `json:"sample_rate"`
		Channels       int    `json:"channels"`
		RFrameRate     string `json:"r_frame_rate"`
		AvgFrameRate   string `json:"avg_frame_rate"`
		ColorTransfer  string `json:"color_transfer"`
		ColorPrimaries string `json:"color_primaries"`
		PixFmt         string `json:"pix_fmt"`
	} `json:"streams"`
}

type mediaInfo struct {
	duration    string
	size        string
	bitrate     string
	videoCodec  string
	audioCodec  string
	width       int
	height      int
	sampleRate  string
	channels    int
	fps         string
	isHDR       bool
	pixFmt      string
	audioTracks int
	subTracks   int
}

func probe(path string) (*mediaInfo, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var result probeResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, err
	}

	info := &mediaInfo{
		duration: result.Format.Duration,
		size:     result.Format.Size,
		bitrate:  result.Format.BitRate,
	}

	for _, stream := range result.Streams {
		switch stream.CodecType {
		case "video":
			if info.videoCodec == "" { // prendre le premier stream vidéo
				info.videoCodec = stream.CodecName
				info.width = stream.Width
				info.height = stream.Height
				info.pixFmt = stream.PixFmt
				// FPS depuis r_frame_rate ou avg_frame_rate
				if stream.RFrameRate != "" && stream.RFrameRate != "0/0" {
					info.fps = stream.RFrameRate
				} else if stream.AvgFrameRate != "" {
					info.fps = stream.AvgFrameRate
				}
				// Détection HDR
				hdrTransfers := map[string]bool{
					"smpte2084": true, "arib-std-b67": true, "smpte428": true,
				}
				hdrPrimaries := map[string]bool{
					"bt2020": true,
				}
				if hdrTransfers[stream.ColorTransfer] || hdrPrimaries[stream.ColorPrimaries] {
					info.isHDR = true
				}
			}
		case "audio":
			info.audioTracks++
			if info.audioCodec == "" { // prendre le premier stream audio
				info.audioCodec = stream.CodecName
				info.sampleRate = stream.SampleRate
				info.channels = stream.Channels
			}
		case "subtitle":
			info.subTracks++
		}
	}

	return info, nil
}

// Génère une preview avec chafa pour une image
func previewImage(path string, width, height int) string {
	cmd := exec.Command("chafa",
		"--size", fmt.Sprintf("%dx%d", width, height),
		"--format", "symbols",
		path,
	)

	output, err := cmd.Output()
	if err != nil {
		return dirStyle.Render("Preview non disponible")
	}

	return string(output)
}

// Génère une preview avec chafa pour une vidéo (extrait une frame)
func previewVideo(path string, width, height int) string {
	// Extraire une frame avec ffmpeg et la passer à chafa
	ffmpeg := exec.Command("ffmpeg",
		"-ss", "00:00:01",
		"-i", path,
		"-frames:v", "1",
		"-f", "image2pipe",
		"-vcodec", "png",
		"-",
	)

	chafa := exec.Command("chafa",
		"--size", fmt.Sprintf("%dx%d", width, height),
		"--format", "symbols",
		"-",
	)

	// Pipe ffmpeg stdout vers chafa stdin
	pipe, err := ffmpeg.StdoutPipe()
	if err != nil {
		return dirStyle.Render("Preview non disponible")
	}
	chafa.Stdin = pipe

	if err := ffmpeg.Start(); err != nil {
		return dirStyle.Render("Preview non disponible")
	}

	output, err := chafa.Output()
	if err != nil {
		ffmpeg.Wait()
		return dirStyle.Render("Preview non disponible")
	}

	ffmpeg.Wait()
	return string(output)
}

// Génère une preview waveform avec ffmpeg pour un fichier audio
func previewAudio(path string, width, height int) string {
	// Générer une image waveform avec ffmpeg et la passer à chafa
	// Dimensions de l'image source (plus grande pour meilleure qualité)
	imgWidth := width * 8
	imgHeight := height * 8

	ffmpeg := exec.Command("ffmpeg",
		"-i", path,
		"-filter_complex", fmt.Sprintf("showwavespic=s=%dx%d:colors=0x7C3AED", imgWidth, imgHeight),
		"-frames:v", "1",
		"-f", "image2pipe",
		"-vcodec", "png",
		"-",
	)

	chafa := exec.Command("chafa",
		"--size", fmt.Sprintf("%dx%d", width, height),
		"--format", "symbols",
		"-",
	)

	// Pipe ffmpeg stdout vers chafa stdin
	pipe, err := ffmpeg.StdoutPipe()
	if err != nil {
		return dirStyle.Render("Preview non disponible")
	}
	chafa.Stdin = pipe

	if err := ffmpeg.Start(); err != nil {
		return dirStyle.Render("Preview non disponible")
	}

	output, err := chafa.Output()
	if err != nil {
		ffmpeg.Wait()
		return dirStyle.Render("Preview non disponible")
	}

	ffmpeg.Wait()
	return string(output)
}

func buildOutputPath(inputPath string, p preset) string {
	dir := filepath.Dir(inputPath)
	base := filepath.Base(inputPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	return filepath.Join(dir, name+"_converted"+p.extension)
}

func buildCommand(input, output string, p preset) []string {
	args := []string{"-i", input}
	args = append(args, p.args...)
	args = append(args, "-y", output)
	return args
}

type model struct {
	view         view
	cursor       int
	entries      []entry
	dir          string
	info         *mediaInfo
	preview      string
	err          string
	message      string
	presetCursor int
	selectedFile string
	outputPath   string
	commandArgs  []string
	width        int
	height       int
	scroll       int
	showHidden   bool
	focusedPanel int  // 0 = gauche (fichiers), 1 = droite (infos)
	useNerdFonts bool // utiliser les icônes Nerd Font
}

func initialModel() model {
	dir, _ := os.Getwd()
	cfg := loadConfig()
	return model{
		view:         viewFiles,
		dir:          dir,
		showHidden:   false,
		entries:      listFiles(dir, false),
		useNerdFonts: cfg.UseNerdFonts,
	}
}

func listFiles(dir string, showHidden bool) []entry {
	items, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var entries []entry

	if dir != "/" {
		entries = append(entries, entry{name: "..", isDir: true})
	}

	for _, item := range items {
		name := item.Name()
		// Masquer les fichiers/dossiers cachés si showHidden est false
		if !showHidden && strings.HasPrefix(name, ".") {
			continue
		}
		entries = append(entries, entry{
			name:      name,
			isDir:     item.IsDir(),
			mediaType: getMediaType(name),
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].name == ".." {
			return true
		}
		if entries[j].name == ".." {
			return false
		}
		if entries[i].isDir != entries[j].isDir {
			return entries[i].isDir
		}
		return strings.ToLower(entries[i].name) < strings.ToLower(entries[j].name)
	})

	return entries
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
		m.preview = previewImage(path, previewWidth, previewHeight)
	case mediaTypeVideo:
		m.preview = previewVideo(path, previewWidth, previewHeight)
	case mediaTypeAudio:
		m.preview = previewAudio(path, previewWidth, previewHeight)
	default:
		m.preview = ""
	}
}

// Charge automatiquement les infos et la preview du fichier sélectionné
// Pour les vidéos : infos auto, preview manuelle (touche p)
// Pour les images/audio : infos ET preview auto
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

	// Charger les infos
	info, err := probe(path)
	if err != nil {
		m.info = nil
	} else {
		m.info = info
	}

	// Charger la preview automatiquement sauf pour les vidéos
	if selected.mediaType == mediaTypeImage {
		m.preview = previewImage(path, previewWidth, previewHeight)
	} else {
		m.preview = ""
	}
}

func (m model) updateFiles(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	listHeight := m.height - 8
	previewWidth := m.width/3 - 6
	previewHeight := (m.height - 16) / 2

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			m.updateScroll(listHeight)
			m.loadCurrentFileInfo(previewWidth, previewHeight)
		}
	case "down", "j":
		if m.cursor < len(m.entries)-1 {
			m.cursor++
			m.updateScroll(listHeight)
			m.loadCurrentFileInfo(previewWidth, previewHeight)
		}
	case "enter", "l":
		if len(m.entries) > 0 {
			selected := m.entries[m.cursor]
			if selected.isDir {
				if selected.name == ".." {
					m.dir = filepath.Dir(m.dir)
				} else {
					m.dir = filepath.Join(m.dir, selected.name)
				}
				m.entries = listFiles(m.dir, m.showHidden)
				m.cursor = 0
				m.scroll = 0
				m.loadCurrentFileInfo(previewWidth, previewHeight)
			}
		}
	case "backspace", "h":
		if m.dir != "/" {
			m.dir = filepath.Dir(m.dir)
			m.entries = listFiles(m.dir, m.showHidden)
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
		m.loadPreview(previewWidth, previewHeight)
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
		m.entries = listFiles(m.dir, m.showHidden)
		m.cursor = 0
		m.scroll = 0
		m.loadCurrentFileInfo(previewWidth, previewHeight)
	case "tab":
		m.focusedPanel = (m.focusedPanel + 1) % 2
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
			m.message = "Conversion terminée !"
			m.entries = listFiles(m.dir, m.showHidden)
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

	// Calculer les dimensions
	leftWidth := max(30, m.width/4)
	rightWidth := m.width - leftWidth - 4
	listHeight := m.height - 3

	if leftWidth < 20 {
		leftWidth = 20
	}
	if rightWidth < 20 {
		rightWidth = 20
	}
	if listHeight < 5 {
		listHeight = 5
	}

	// Panel gauche : fichiers
	var leftContent strings.Builder
	for i, e := range m.entries {
		if i < m.scroll || i >= m.scroll+listHeight {
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
		maxNameLen := leftWidth - 6
		if len(name) > maxNameLen {
			name = name[:maxNameLen-3] + "..."
		}

		line := fmt.Sprintf("%s %s", icon, name)

		if i == m.cursor {
			line = selectedStyle.Render(line)
		} else {
			line = style.Render(line)
		}

		leftContent.WriteString(line + "\n")
	}
	leftContentStr := strings.TrimSuffix(leftContent.String(), "\n")

	// Panel droit : infos + preview
	var rightContent strings.Builder

	// Preview
	if m.preview != "" {
		rightContent.WriteString(panelTitleStyle.Render("Preview") + "\n")
		rightContent.WriteString(m.preview + "\n")
	}

	// Infos
	rightContent.WriteString(panelTitleStyle.Render("Infos") + "\n\n")

	// Déterminer le type de média actuel
	var currentMediaType mediaType
	if len(m.entries) > 0 && m.cursor < len(m.entries) {
		currentMediaType = m.entries[m.cursor].mediaType
	}

	if m.info != nil {
		// Taille (toujours affichée)
		if m.info.size != "" {
			rightContent.WriteString(infoLabelStyle.Render("Taille:"))
			rightContent.WriteString(infoValueStyle.Render(formatSize(m.info.size)) + "\n")
		}

		if currentMediaType == mediaTypeImage {
			// === AFFICHAGE IMAGE ===
			if m.info.width > 0 && m.info.height > 0 {
				rightContent.WriteString(infoLabelStyle.Render("Dimensions:"))
				rightContent.WriteString(infoValueStyle.Render(formatResolution(m.info.width, m.info.height)) + "\n")
			}
			if m.info.pixFmt != "" {
				rightContent.WriteString(infoLabelStyle.Render("Format:"))
				rightContent.WriteString(infoValueStyle.Render(formatPixelFormat(m.info.pixFmt)) + "\n")
			}
			// BPP (bits per pixel)
			if m.info.size != "" && m.info.width > 0 && m.info.height > 0 {
				rightContent.WriteString(infoLabelStyle.Render("BPP:"))
				rightContent.WriteString(infoValueStyle.Render(formatBPP(m.info.size, m.info.width, m.info.height)) + "\n")
			}
			// Transparence
			rightContent.WriteString(infoLabelStyle.Render("Transparence:"))
			if hasAlpha(m.info.pixFmt) {
				rightContent.WriteString(infoValueStyle.Render("Oui") + "\n")
			} else {
				rightContent.WriteString(infoValueStyle.Render("Non") + "\n")
			}
		} else {
			// === AFFICHAGE VIDÉO/AUDIO ===
			if m.info.duration != "" {
				rightContent.WriteString(infoLabelStyle.Render("Durée:"))
				rightContent.WriteString(infoValueStyle.Render(formatDuration(m.info.duration)) + "\n")
			}
			if m.info.bitrate != "" {
				rightContent.WriteString(infoLabelStyle.Render("Bitrate:"))
				rightContent.WriteString(infoValueStyle.Render(formatBitrate(m.info.bitrate)) + "\n")
			}

			// Section Vidéo
			if m.info.videoCodec != "" && currentMediaType == mediaTypeVideo {
				rightContent.WriteString("\n" + panelTitleStyle.Render("Vidéo") + "\n")
				rightContent.WriteString(infoLabelStyle.Render("Codec:"))
				rightContent.WriteString(infoValueStyle.Render(m.info.videoCodec) + "\n")
				rightContent.WriteString(infoLabelStyle.Render("Résolution:"))
				rightContent.WriteString(infoValueStyle.Render(formatResolution(m.info.width, m.info.height)) + "\n")
				if m.info.fps != "" {
					rightContent.WriteString(infoLabelStyle.Render("FPS:"))
					rightContent.WriteString(infoValueStyle.Render(formatFPS(m.info.fps)) + "\n")
				}
				if m.info.isHDR {
					rightContent.WriteString(infoLabelStyle.Render("HDR:"))
					rightContent.WriteString(infoValueStyle.Render("Oui") + "\n")
				}
			}

			// Section Audio
			if m.info.audioCodec != "" {
				rightContent.WriteString("\n" + panelTitleStyle.Render("Audio") + "\n")
				rightContent.WriteString(infoLabelStyle.Render("Codec:"))
				rightContent.WriteString(infoValueStyle.Render(m.info.audioCodec) + "\n")
				if m.info.sampleRate != "" {
					rightContent.WriteString(infoLabelStyle.Render("Sample:"))
					rightContent.WriteString(infoValueStyle.Render(m.info.sampleRate+" Hz") + "\n")
				}
				if m.info.channels > 0 {
					rightContent.WriteString(infoLabelStyle.Render("Canaux:"))
					rightContent.WriteString(infoValueStyle.Render(formatChannels(m.info.channels)) + "\n")
				}
			}

			// Pistes additionnelles
			if m.info.audioTracks > 1 || m.info.subTracks > 0 {
				rightContent.WriteString("\n" + panelTitleStyle.Render("Pistes") + "\n")
				if m.info.audioTracks > 1 {
					rightContent.WriteString(infoLabelStyle.Render("Audio:"))
					rightContent.WriteString(infoValueStyle.Render(fmt.Sprintf("%d pistes", m.info.audioTracks)) + "\n")
				}
				if m.info.subTracks > 0 {
					rightContent.WriteString(infoLabelStyle.Render("Sous-titres:"))
					rightContent.WriteString(infoValueStyle.Render(fmt.Sprintf("%d pistes", m.info.subTracks)) + "\n")
				}
			}
		}
	} else {
		rightContent.WriteString(dirStyle.Render("Sélectionnez un média"))
	}

	// Construire les panels avec couleurs selon le focus
	focusedColor := lipgloss.Color("#7C3AED")
	unfocusedColor := lipgloss.Color("#4B5563")

	var leftPanelStyle, rightPanelStyle lipgloss.Style
	var leftBorderColor, rightBorderColor lipgloss.Color

	if m.focusedPanel == 0 {
		leftPanelStyle = panelStyle
		rightPanelStyle = panelStyleUnfocused
		leftBorderColor = focusedColor
		rightBorderColor = unfocusedColor
	} else {
		leftPanelStyle = panelStyleUnfocused
		rightPanelStyle = panelStyle
		leftBorderColor = unfocusedColor
		rightBorderColor = focusedColor
	}

	leftPanel := leftPanelStyle.Width(leftWidth).Height(listHeight).Render(leftContentStr)
	leftPanelWithBorder := lipgloss.JoinVertical(lipgloss.Left, createTopBorder(leftWidth, "Files explorer", leftBorderColor), leftPanel)
	rightPanel := rightPanelStyle.Width(rightWidth).Height(listHeight).Render(rightContent.String())
	rightPanelWithBorder := lipgloss.JoinVertical(lipgloss.Left, createTopBorder(rightWidth, "Infos", rightBorderColor), rightPanel)

	// Joindre horizontalement
	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanelWithBorder, "", rightPanelWithBorder)
	s.WriteString(panels + "\n")

	// Messages
	if m.err != "" {
		s.WriteString(errorStyle.Render("⚠️  "+m.err) + "\n")
	}
	if m.message != "" {
		s.WriteString(successStyle.Render("✅ "+m.message) + "\n")
	}

	// Help bar
	s.WriteString(helpStyle.Render("j/k: naviguer • o: ouvrir • p: preview • c: convertir • q: quitter"))

	return s.String()
}

func (m model) viewPresets() string {
	var s strings.Builder

	header := titleStyle.Render("🎬 ffterm") + " " + dirStyle.Render("Conversion")
	s.WriteString(header + "\n\n")

	s.WriteString(infoLabelStyle.Render("Fichier:"))
	s.WriteString(infoValueStyle.Render(filepath.Base(m.selectedFile)) + "\n\n")

	panelWidth := m.width - 4
	listHeight := m.height - 10

	var content strings.Builder
	for i, p := range presets {
		if i >= listHeight {
			break
		}

		if i == m.presetCursor {
			line := selectedStyle.Render("> " + p.name)
			content.WriteString(line + "  " + presetDescStyle.Render(p.description) + "\n")
		} else {
			content.WriteString("  " + p.name + "  " + presetDescStyle.Render(p.description) + "\n")
		}
	}

	panel := panelStyle.Width(panelWidth).Render(content.String())
	s.WriteString(panel + "\n")

	s.WriteString(helpStyle.Render("j/k: naviguer • enter: sélectionner • backspace: retour • q: quitter"))

	return s.String()
}

func (m model) viewConfirm() string {
	var s strings.Builder

	header := titleStyle.Render("🎬 ffterm") + " " + dirStyle.Render("Confirmer")
	s.WriteString(header + "\n\n")

	panelWidth := m.width - 4

	var content strings.Builder
	content.WriteString(infoLabelStyle.Render("Input:"))
	content.WriteString(infoValueStyle.Render(m.selectedFile) + "\n\n")
	content.WriteString(infoLabelStyle.Render("Output:"))
	content.WriteString(successStyle.Render(m.outputPath) + "\n\n")
	content.WriteString(dirStyle.Render("Commande:") + "\n")
	content.WriteString(commandStyle.Render("ffmpeg "+strings.Join(m.commandArgs, " ")) + "\n")

	panel := panelStyle.Width(panelWidth).Render(content.String())
	s.WriteString(panel + "\n")

	s.WriteString(helpStyle.Render("enter/y: exécuter • n/backspace: annuler • q: quitter"))

	return s.String()
}

func formatSize(sizeStr string) string {
	var size float64
	fmt.Sscanf(sizeStr, "%f", &size)

	if size < 1024 {
		return fmt.Sprintf("%.0f B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", size/1024)
	}
	if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", size/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", size/(1024*1024*1024))
}

func formatBitrate(bitrateStr string) string {
	var bitrate float64
	fmt.Sscanf(bitrateStr, "%f", &bitrate)

	if bitrate < 1000 {
		return fmt.Sprintf("%.0f bps", bitrate)
	}
	if bitrate < 1000000 {
		return fmt.Sprintf("%.0f kbps", bitrate/1000)
	}
	return fmt.Sprintf("%.1f Mbps", bitrate/1000000)
}

func formatDuration(durationStr string) string {
	var seconds float64
	fmt.Sscanf(durationStr, "%f", &seconds)

	if seconds < 60 {
		return fmt.Sprintf("%.0fs", seconds)
	}

	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%02dm%02ds", hours, minutes, secs)
	}
	return fmt.Sprintf("%dm%02ds", minutes, secs)
}

func formatFPS(fpsStr string) string {
	// FPS est souvent au format "30000/1001" ou "25/1"
	var num, den int
	if n, _ := fmt.Sscanf(fpsStr, "%d/%d", &num, &den); n == 2 && den > 0 {
		fps := float64(num) / float64(den)
		if fps == float64(int(fps)) {
			return fmt.Sprintf("%.0f", fps)
		}
		return fmt.Sprintf("%.2f", fps)
	}
	return fpsStr
}

func formatResolution(width, height int) string {
	if width == 0 || height == 0 {
		return ""
	}

	// Calcul du ratio
	gcd := func(a, b int) int {
		for b != 0 {
			a, b = b, a%b
		}
		return a
	}
	g := gcd(width, height)
	ratioW, ratioH := width/g, height/g

	// Simplifier les ratios courants
	ratio := fmt.Sprintf("%d:%d", ratioW, ratioH)
	ratioMap := map[string]string{
		"16:9": "16:9", "4:3": "4:3", "21:9": "21:9",
		"64:27": "21:9", "43:18": "21:9", "12:5": "21:9",
		"8:5": "16:10", "16:10": "16:10",
		"3:2": "3:2", "1:1": "1:1",
	}
	if simplified, ok := ratioMap[ratio]; ok {
		ratio = simplified
	}

	return fmt.Sprintf("%d×%d (%s)", width, height, ratio)
}

func formatChannels(channels int) string {
	switch channels {
	case 1:
		return "Mono"
	case 2:
		return "Stéréo"
	case 6:
		return "5.1"
	case 8:
		return "7.1"
	default:
		return fmt.Sprintf("%d canaux", channels)
	}
}

func formatPixelFormat(pixFmt string) string {
	// Mapping des formats pixel vers noms lisibles
	formats := map[string]string{
		"rgb24":     "RGB",
		"rgba":      "RGBA",
		"rgb48":     "RGB 16-bit",
		"rgba64":    "RGBA 16-bit",
		"bgr24":     "BGR",
		"bgra":      "BGRA",
		"gray":      "Grayscale",
		"gray16":    "Grayscale 16-bit",
		"ya8":       "Grayscale + Alpha",
		"yuv420p":   "YUV 4:2:0",
		"yuv422p":   "YUV 4:2:2",
		"yuv444p":   "YUV 4:4:4",
		"yuvj420p":  "YUV 4:2:0",
		"yuvj422p":  "YUV 4:2:2",
		"yuvj444p":  "YUV 4:4:4",
		"pal8":      "Palette 8-bit",
	}
	if name, ok := formats[pixFmt]; ok {
		return name
	}
	return pixFmt
}

func formatBPP(sizeStr string, width, height int) string {
	var size float64
	fmt.Sscanf(sizeStr, "%f", &size)

	pixels := float64(width * height)
	if pixels == 0 {
		return "N/A"
	}

	bpp := (size * 8) / pixels

	if bpp < 1 {
		return fmt.Sprintf("%.2f (très compressé)", bpp)
	} else if bpp < 4 {
		return fmt.Sprintf("%.2f (compressé)", bpp)
	} else if bpp < 16 {
		return fmt.Sprintf("%.2f (normal)", bpp)
	}
	return fmt.Sprintf("%.2f (peu compressé)", bpp)
}

func hasAlpha(pixFmt string) bool {
	alphaFormats := map[string]bool{
		"rgba": true, "bgra": true, "argb": true, "abgr": true,
		"rgba64": true, "bgra64": true,
		"ya8": true, "ya16": true,
		"pal8": true, // peut avoir de la transparence
		"yuva420p": true, "yuva422p": true, "yuva444p": true,
	}
	return alphaFormats[pixFmt]
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
