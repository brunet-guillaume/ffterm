package main

import (
	"fmt"
	"os/exec"
)

func previewImage(path string, width, height int) string {
	cmd := exec.Command("chafa",
		"--size", fmt.Sprintf("%dx%d", width, height),
		"--format", "symbols",
		path,
	)

	output, err := cmd.Output()
	if err != nil {
		return dirStyle.Render("\n\n\n\nPreview unavailable")
	}

	return string(output)
}

func previewVideo(path string, width, height int) string {
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

	pipe, err := ffmpeg.StdoutPipe()
	if err != nil {
		return dirStyle.Render("\n\n\n\nPreview unavailable")
	}
	chafa.Stdin = pipe

	if err := ffmpeg.Start(); err != nil {
		return dirStyle.Render("\n\n\n\nPreview unavailable")
	}

	output, err := chafa.Output()
	if err != nil {
		ffmpeg.Wait()
		return dirStyle.Render("\n\n\n\nPreview unavailable")
	}

	ffmpeg.Wait()
	return string(output)
}

func previewAudio(path string, width, height int) string {
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

	pipe, err := ffmpeg.StdoutPipe()
	if err != nil {
		return dirStyle.Render("\n\n\n\nPreview unavailable")
	}
	chafa.Stdin = pipe

	if err := ffmpeg.Start(); err != nil {
		return dirStyle.Render("\n\n\n\nPreview unavailable")
	}

	output, err := chafa.Output()
	if err != nil {
		ffmpeg.Wait()
		return dirStyle.Render("\n\n\n\nPreview unavailable")
	}

	ffmpeg.Wait()
	return string(output)
}
