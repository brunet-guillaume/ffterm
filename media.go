package main

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
)

type mediaType int

const (
	mediaTypeNone mediaType = iota
	mediaTypeVideo
	mediaTypeAudio
	mediaTypeImage
)

type entry struct {
	name      string
	isDir     bool
	mediaType mediaType
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
			if info.videoCodec == "" {
				info.videoCodec = stream.CodecName
				info.width = stream.Width
				info.height = stream.Height
				info.pixFmt = stream.PixFmt
				if stream.RFrameRate != "" && stream.RFrameRate != "0/0" {
					info.fps = stream.RFrameRate
				} else if stream.AvgFrameRate != "" {
					info.fps = stream.AvgFrameRate
				}
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
			if info.audioCodec == "" {
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
