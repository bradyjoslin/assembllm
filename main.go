package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/huh"
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
}

var (
	appCfg AppConfig

	logLevel = extism.LogLevelError

	rootCmd = &cobra.Command{
		Use:           "assembllm [prompt]",
		RunE:          runCommand,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	configPath = "config.yaml"
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

// Initializes the flags for the root command
func initializeFlags() {
	flags := rootCmd.Flags()
	flags.StringVarP(&appCfg.Name, "plugin", "p", "openai", "The name of the plugin to use")
	flags.StringVarP(&appCfg.Model, "model", "m", "", "The name of the model to use")
	flags.BoolVarP(&appCfg.ChooseAIModel, "choose-model", "c", false, "Choose the model to use")
	flags.StringVarP(&appCfg.Temperature, "temperature", "t", "", "The temperature to use")
	flags.StringVarP(&appCfg.Role, "role", "r", "", "The role to use")
	flags.BoolVarP(&appCfg.Raw, "raw", "", false, "Raw output without formatting")
	flags.SortFlags = false
}

// Generates a prompt for the chat completions
// If there is input from stdin, it will be included in the prompt
// If the user specified a prompt as an argument, it will be included in the prompt
// If there is no prompt, the user will be prompted to enter one
func generatePrompt(args []string, raw bool) string {
	var prompt string
	stdInStats, _ := os.Stdin.Stat()

	if stdInStats.Size() > 0 {
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

	res, err := pluginCfg.generateResponse(prompt, appCfg.Raw)
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
