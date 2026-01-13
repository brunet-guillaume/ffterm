package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type config struct {
	UseNerdFonts bool `json:"use_nerd_fonts"`
}

func getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ffterm", "config.json")
}

func loadConfig() config {
	cfg := config{UseNerdFonts: true}

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
