package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Port string `json:"port"`
	IP   string `json:"ip"`
}

func getConfigPath() string {
	var configDir string
	
	if os.Getenv("APPDATA") != "" {
		configDir = filepath.Join(os.Getenv("APPDATA"), "QAA-AirType-Go")
	} else {
		homeDir, _ := os.UserHomeDir()
		configDir = filepath.Join(homeDir, ".config", "qaa-airtype-go")
	}
	
	os.MkdirAll(configDir, 0755)
	return filepath.Join(configDir, "config.json")
}

func Load() Config {
	config := Config{Port: "5000"}
	
	data, err := os.ReadFile(getConfigPath())
	if err != nil {
		return config
	}
	
	json.Unmarshal(data, &config)
	return config
}

func Save(config Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(getConfigPath(), data, 0644)
}