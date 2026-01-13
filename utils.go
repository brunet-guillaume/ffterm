package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
)

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
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	return cmd.Start()
}

func listFiles(dir string, showHidden bool, showOnlyMedia bool) []entry {
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
		if !showHidden && strings.HasPrefix(name, ".") {
			continue
		}
		mediaType := getMediaType(name)
		isDir := item.IsDir()
		if showOnlyMedia && !isDir && mediaType == mediaTypeNone {
			continue
		}
		entries = append(entries, entry{
			name:      name,
			isDir:     isDir,
			mediaType: mediaType,
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
