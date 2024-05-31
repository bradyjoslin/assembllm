package main

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	extism "github.com/extism/go-sdk"
	"github.com/spf13/cobra"
)

type AppConfig struct {
	Name          string
	Model         string
	ChooseAIModel bool
	Temperature   string
	Role          string
	Raw           bool
	Version       bool
}

const (
	configFileName = "config.yaml"
	version        = "0.1.0"
)

var (
	appCfg AppConfig

	//go:embed config.yaml
	defaultConfig []byte

	configPath string

	logLevel = extism.LogLevelOff

	appName = filepath.Base(os.Args[0])
	rootCmd = &cobra.Command{
		Use:           appName + " [prompt]",
		Short:         "A WASM plug-in based CLI for AI chat completions",
		RunE:          runCommand,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
)

// Gets the available models from the completions plugin and prompts the user to choose one
func chooseModel(pluginCfg CompletionPluginConfig) (string, error) {
	modelNames, err := pluginCfg.getModels()
	if err != nil {
		return "", fmt.Errorf("failed to get models: %v", err)
	}

	var opts []huh.Option[string]
	for _, model := range modelNames {
		opts = append(opts, huh.Option[string]{
			Key:   model,
			Value: model,
		})
	}

	var model string
	huh.NewSelect[string]().
		Title("Choose a model:").
		Options(opts...).
		Value(&model).
		Run()

	return model, nil
}

func init() {
	// Get the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Unable to get user's home directory: %v", err)
	}

	// Define the path to the configuration file
	configPath = filepath.Join(homeDir, "."+appName, configFileName)

	// Check if the configuration file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Ensure the directory exists
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			log.Fatalf("Unable to create directory for config file: %v", err)
		}

		// Write the default configuration to the new configuration file
		err = os.WriteFile(configPath, defaultConfig, 0644)
		if err != nil {
			log.Fatalf("Unable to write default config to file: %v", err)
		}
	}
}

// Initializes the flags for the root command
func initializeFlags() {
	flags := rootCmd.Flags()
	flags.StringVarP(&appCfg.Name, "plugin", "p", "openai", "The name of the plugin to use")
	flags.StringVarP(&appCfg.Model, "model", "m", "", "The name of the model to use")
	flags.BoolVarP(&appCfg.ChooseAIModel, "choose-model", "c", false, "Choose the model to use")
	flags.StringVarP(&appCfg.Temperature, "temperature", "t", "", "The temperature to use")
	flags.StringVarP(&appCfg.Role, "role", "r", "", "The role to use")
	flags.BoolVarP(&appCfg.Raw, "raw", "", false, "Raw output without formatting")
	flags.BoolVarP(&appCfg.Version, "version", "v", false, "Print the version")
	flags.SortFlags = false
}

// Generates a prompt for the chat completions
// If there is input from stdin, it will be included in the prompt
// If the user specified a prompt as an argument, it will be included in the prompt
// If there is no prompt, the user will be prompted to enter one
func generatePrompt(args []string, raw bool) string {
	var prompt string
	stdInStats, _ := os.Stdin.Stat()

	if (stdInStats.Mode() & os.ModeCharDevice) == 0 {
		reader := bufio.NewReader(os.Stdin)
		s, _ := io.ReadAll(reader)
		prompt += string(s)
	}

	if len(args) == 1 {
		prompt += args[0]
	}

	if prompt == "" {
		huh.NewInput().
			Title("Enter a chat completions prompt:").
			Value(&prompt).
			Run()
	}

	if raw {
		prompt += "\nomit any markdown formatting in response"
	}

	return prompt
}

// Overrides the plugin config with the user flags
func overridePluginConfigWithUserFlags(appConfig AppConfig, pluginConfig CompletionPluginConfig) CompletionPluginConfig {
	if appConfig.Model != "" {
		pluginConfig.Model = appConfig.Model
	}

	if appConfig.Temperature != "" {
		pluginConfig.Temperature = appConfig.Temperature
	}

	if appConfig.Role != "" {
		pluginConfig.Role = appConfig.Role
	}

	return pluginConfig
}

func runCommand(cmd *cobra.Command, args []string) error {
	if appCfg.Version {
		fmt.Println(appName + " " + version)
		return nil
	}

	pluginCfg, err := getPluginConfig(appCfg.Name, configPath)
	if err != nil {
		return err
	}

	pluginCfg = overridePluginConfigWithUserFlags(appCfg, pluginCfg)

	if appCfg.ChooseAIModel {
		pluginCfg.Model, err = chooseModel(pluginCfg)
		if err != nil {
			return err
		}
	}

	prompt := generatePrompt(args, appCfg.Raw)

	ctx, cancel := context.WithCancel(context.Background())

	var res string

	action := func() {
		res, err = pluginCfg.generateResponse(prompt, appCfg.Raw)
		cancel()
	}

	go action()
	_ = spinner.New().Title("Generating...").Context(ctx).Run()

	if err != nil {
		return err
	}

	fmt.Print(res)
	return nil
}

func main() {
	initializeFlags()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)

}
