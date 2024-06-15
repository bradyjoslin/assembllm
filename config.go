package main

import (
	_ "embed"
	"log"
	"os"
	"path/filepath"

	extism "github.com/extism/go-sdk"
)

const configFileName = "config.yaml"

var (
	//go:embed config.yaml
	defaultConfig []byte
)

type CompletionPluginConfig struct {
	Name        string `yaml:"name"`
	Source      string `yaml:"source"`
	Hash        string `yaml:"hash"`
	APIKey      string `yaml:"apiKey"`
	AccountId   string `yaml:"accountId"`
	URL         string `yaml:"url"`
	Model       string `yaml:"model"`
	Temperature string `yaml:"temperature"`
	Role        string `yaml:"role"`
	Wasi        bool   `yaml:"wasi"`
	LogLevel    extism.LogLevel
}

type CompletionPluginConfigs struct {
	Plugins []CompletionPluginConfig `yaml:"completion-plugins"`
}

type Config struct {
	CompletionPluginConfigs CompletionPluginConfigs `yaml:"completion-plugins"`
}

func createConfig(configPath string) {
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		log.Fatalf("Unable to create directory for config file: %v", err)
	}

	// Write the default configuration to the new configuration file
	err := os.WriteFile(configPath, defaultConfig, 0600)
	if err != nil {
		log.Fatalf("Unable to write default config to file: %v", err)
	}
}

func readConfig(configPath string) []byte {
	// Read the existing config file
	configData, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Unable to read config file: %v", err)
	}
	return configData
}

func writeConfig(configDataUpdates []byte, configPath string) {
	// Write the updated config back to the file
	err := os.WriteFile(configPath, configDataUpdates, 0600)
	if err != nil {
		log.Fatalf("Unable to write updated config to file: %v", err)
	}
}

func getConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Unable to get user config directory: %v", err)
	}
	configPath := filepath.Join(configDir, "."+appName, configFileName)
	return configPath
}

func setupConfig() {
	configPath := getConfigPath()

	// Check if the configuration file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		createConfig(configPath)
	} else {
		configData := readConfig(configPath)

		configDataUpdates := updatePlugins(configData)

		writeConfig(configDataUpdates, configPath)
	}
}
