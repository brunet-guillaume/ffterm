package main

import (
	"path/filepath"
	"strings"
)

type preset struct {
	name        string
	description string
	extension   string
	args        []string
}

var presets = []preset{
	{
		name:        "MP4 (H.264)",
		description: "Universal format",
		extension:   ".mp4",
		args:        []string{"-c:v", "libx264", "-preset", "medium", "-crf", "23", "-c:a", "aac", "-b:a", "128k"},
	},
	{
		name:        "MP4 (H.265)",
		description: "Better compression",
		extension:   ".mp4",
		args:        []string{"-c:v", "libx265", "-preset", "medium", "-crf", "28", "-c:a", "aac", "-b:a", "128k"},
	},
	{
		name:        "WebM (VP9)",
		description: "Web optimized",
		extension:   ".webm",
		args:        []string{"-c:v", "libvpx-vp9", "-crf", "30", "-b:v", "0", "-c:a", "libopus", "-b:a", "128k"},
	},
	{
		name:        "MP3 320kbps",
		description: "HQ Audio",
		extension:   ".mp3",
		args:        []string{"-vn", "-c:a", "libmp3lame", "-b:a", "320k"},
	},
	{
		name:        "AAC 256kbps",
		description: "Apple Audio",
		extension:   ".m4a",
		args:        []string{"-vn", "-c:a", "aac", "-b:a", "256k"},
	},
	{
		name:        "FLAC",
		description: "Lossless audio",
		extension:   ".flac",
		args:        []string{"-vn", "-c:a", "flac"},
	},
	{
		name:        "GIF",
		description: "Animated 480px",
		extension:   ".gif",
		args:        []string{"-vf", "fps=10,scale=480:-1:flags=lanczos", "-loop", "0"},
	},
	{
		name:        "Twitter",
		description: "720p optimized",
		extension:   ".mp4",
		args:        []string{"-c:v", "libx264", "-preset", "slow", "-crf", "24", "-vf", "scale=-2:720", "-c:a", "aac", "-b:a", "128k"},
	},
	{
		name:        "Light compression",
		description: "Reduce ~30%",
		extension:   ".mp4",
		args:        []string{"-c:v", "libx264", "-preset", "medium", "-crf", "26", "-c:a", "aac", "-b:a", "128k"},
	},
	{
		name:        "Heavy compression",
		description: "Reduce ~60%",
		extension:   ".mp4",
		args:        []string{"-c:v", "libx264", "-preset", "slow", "-crf", "32", "-c:a", "aac", "-b:a", "96k"},
	},
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
