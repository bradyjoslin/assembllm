package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
	extism "github.com/extism/go-sdk"
	"gopkg.in/yaml.v3"
)

type Model struct {
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
}

type CompletionsPlugin struct {
	Plugin extism.Plugin
}

// Get the available models from the completions plugin
func (pluginCfg CompletionPluginConfig) getModels() ([]string, error) {
	plugin, err := pluginCfg.createPlugin()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize plugin: %v", err)
	}

	modelNames, err := plugin.getModelNames()
	if err != nil {
		return nil, fmt.Errorf("failed to get models: %v", err)
	}

	return modelNames, nil
}

// Get completions response for the prompt from the completions plugin
func (pluginInfo CompletionPluginConfig) generateResponse(prompt string, raw bool) (string, error) {
	plugin, err := pluginInfo.createPlugin()
	if err != nil {
		return "", fmt.Errorf("failed to initialize plugin: %v", err)
	}

	_, out, err := plugin.completion(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to get completion: %v", err)
	}

	response := string(out)

	if raw {
		return response, nil
	} else {
		formattedResponse, _ := glamour.Render(response, "dark")
		return formattedResponse, nil
	}
}

// Call an exposed Extism function on the completions plugin
func (p *CompletionsPlugin) Call(method string, payload []byte) (uint32, []byte, error) {
	return p.Plugin.Call(method, payload)
}

// Create a new completions extism plugin from the configuration
func (p CompletionPluginConfig) createPlugin() (CompletionsPlugin, error) {
	var wasm extism.Wasm

	if strings.HasPrefix(p.Source, "https://") {
		wasm = extism.WasmUrl{
			Url:  p.Source,
			Hash: p.Hash,
		}
	} else {
		homeDir, _ := os.UserHomeDir()
		p.Source = strings.Replace(p.Source, "~", homeDir, 1)

		if !isFilePath(p.Source) {
			return CompletionsPlugin{}, fmt.Errorf("file not found: %s", p.Source)
		}

		wasm = extism.WasmFile{
			Path: p.Source,
			Hash: p.Hash,
		}
	}

	manifest := extism.Manifest{
		Wasm: []extism.Wasm{
			wasm,
		},
	}

	plugin, err := extism.NewPlugin(
		context.Background(),
		manifest,
		extism.PluginConfig{
			EnableWasi: p.Wasi,
		},
		[]extism.HostFunction{},
	)
	if err != nil {
		return CompletionsPlugin{}, err
	}
	if plugin == nil {
		return CompletionsPlugin{}, fmt.Errorf("plugin is nil")
	}

	plugin.AllowedHosts = []string{p.URL}
	plugin.Config = map[string]string{"api_key": p.APIKey, "model": p.Model, "temperature": p.Temperature, "role": p.Role, "account_id": p.AccountId}

	plugin.SetLogLevel(p.LogLevel)
	plugin.SetLogger(func(level extism.LogLevel, message string) {
		fmt.Printf("[%s] %s\n", level, message)
	})
	return CompletionsPlugin{*plugin}, nil
}

// Get list of supported models
func (plugin *CompletionsPlugin) models() (uint32, []byte, error) {
	return plugin.Call("models", []byte{})
}

// Get completions for the prompt
func (plugin *CompletionsPlugin) completion(prompt string) (uint32, []byte, error) {
	return plugin.Call("completion", []byte(prompt))
}

// Get a plugin configuration from the available plugins
func (plugins CompletionPluginConfigs) getPlugin(pluginName string) (CompletionPluginConfig, error) {
	var pluginInfo CompletionPluginConfig
	for _, p := range plugins.Plugins {
		if p.Name == pluginName {
			pluginInfo = p
			break
		}
	}
	if pluginInfo.Name == "" {
		return CompletionPluginConfig{}, fmt.Errorf("plugin not found: %s", pluginName)
	}
	return pluginInfo, nil
}

// Get the available models from the completions plugin
func (plugin CompletionsPlugin) getModelNames() ([]string, error) {
	_, jsonMs, err := plugin.models()
	if err != nil {
		return nil, fmt.Errorf("failed to get models: %v", err)
	}
	var models []Model
	err = json.Unmarshal([]byte(jsonMs), &models)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal models: %v", err)
	}

	var modelNames []string
	for _, model := range models {
		modelNames = append(modelNames, model.Name)
	}
	return modelNames, nil
}

// Unmarshal the plugin configuration from the yaml file
func (p *CompletionPluginConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawPlugin CompletionPluginConfig
	raw := rawPlugin{}

	if err := unmarshal(&raw); err != nil {
		return err
	}

	raw.APIKey = os.Getenv(raw.APIKey)
	raw.AccountId = os.Getenv(raw.AccountId)
	raw.LogLevel = logLevel

	*p = CompletionPluginConfig(raw)

	return nil
}

// Check if the path is a file
func isFilePath(s string) bool {
	info, err := os.Stat(s)
	return !os.IsNotExist(err) && !info.IsDir()
}

// Loads the available chat completion plugins from a yaml file
func getAvailablePlugins(filename string) (CompletionPluginConfigs, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return CompletionPluginConfigs{}, err
	}

	var completionPluginConfigs CompletionPluginConfigs
	err = yaml.Unmarshal(file, &completionPluginConfigs)
	if err != nil {
		return CompletionPluginConfigs{}, err
	}

	return completionPluginConfigs, nil
}

// Gets the available plugins from the yaml file, then gets the plugin config for the specified plugin
func getPluginConfig(pluginName string, configPath string) (CompletionPluginConfig, error) {
	pluginConfigs, err := getAvailablePlugins(configPath)
	if err != nil {
		return CompletionPluginConfig{}, fmt.Errorf("failed to get config from yaml: %v", err)
	}

	pluginCfg, err := pluginConfigs.getPlugin(pluginName)
	if err != nil {
		return CompletionPluginConfig{}, fmt.Errorf("failed to get plugin info: %v", err)
	}

	return pluginCfg, nil
}
