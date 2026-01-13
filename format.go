package main

import "fmt"

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

	gcd := func(a, b int) int {
		for b != 0 {
			a, b = b, a%b
		}
		return a
	}
	g := gcd(width, height)
	ratioW, ratioH := width/g, height/g

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
		return "Stereo"
	case 6:
		return "5.1"
	case 8:
		return "7.1"
	default:
		return fmt.Sprintf("%d channels", channels)
	}
}

func formatPixelFormat(pixFmt string) string {
	formats := map[string]string{
		"rgb24":    "RGB",
		"rgba":     "RGBA",
		"rgb48":    "RGB 16-bit",
		"rgba64":   "RGBA 16-bit",
		"bgr24":    "BGR",
		"bgra":     "BGRA",
		"gray":     "Grayscale",
		"gray16":   "Grayscale 16-bit",
		"ya8":      "Grayscale + Alpha",
		"yuv420p":  "YUV 4:2:0",
		"yuv422p":  "YUV 4:2:2",
		"yuv444p":  "YUV 4:4:4",
		"yuvj420p": "YUV 4:2:0",
		"yuvj422p": "YUV 4:2:2",
		"yuvj444p": "YUV 4:4:4",
		"pal8":     "Palette 8-bit",
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
		return fmt.Sprintf("%.2f (highly compressed)", bpp)
	} else if bpp < 4 {
		return fmt.Sprintf("%.2f (compressed)", bpp)
	} else if bpp < 16 {
		return fmt.Sprintf("%.2f (normal)", bpp)
	}
	return fmt.Sprintf("%.2f (lightly compressed)", bpp)
}

func hasAlpha(pixFmt string) bool {
	alphaFormats := map[string]bool{
		"rgba": true, "bgra": true, "argb": true, "abgr": true,
		"rgba64": true, "bgra64": true,
		"ya8": true, "ya16": true,
		"pal8":     true,
		"yuva420p": true, "yuva422p": true, "yuva444p": true,
	}
	return alphaFormats[pixFmt]
}
