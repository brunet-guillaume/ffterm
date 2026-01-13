package main

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
