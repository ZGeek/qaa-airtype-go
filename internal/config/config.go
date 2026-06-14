package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Port              string  `json:"port"`
	IP                string  `json:"ip"`
	Sensitivity       float64 `json:"sensitivity"`
	ScrollSensitivity float64 `json:"scrollSensitivity,omitempty"`
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
	config := Config{Port: "5000", Sensitivity: 1.5}

	data, err := os.ReadFile(getConfigPath())
	if err != nil {
		return config
	}

	json.Unmarshal(data, &config)
	if config.Port == "" {
		config.Port = "5000"
	}
	if config.Sensitivity <= 0 && config.ScrollSensitivity > 0 {
		config.Sensitivity = config.ScrollSensitivity
	}
	config.ScrollSensitivity = 0
	if config.Sensitivity <= 0 {
		config.Sensitivity = 1.5
	} else if config.Sensitivity > 5 {
		config.Sensitivity = 5
	}
	return config
}

func Save(config Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(getConfigPath(), data, 0644)
}
