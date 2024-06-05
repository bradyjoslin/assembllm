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
	"github.com/charmbracelet/lipgloss"
	extism "github.com/extism/go-sdk"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Name          string
	Model         string
	ChooseAIModel bool
	ChoosePlugin  bool
	Temperature   string
	Role          string
	Raw           bool
	Version       bool
	TasksPath     string
}

type Tasks struct {
	Tasks []Task `yaml:"tasks"`
}

type Task struct {
	Name        string `yaml:"name"`
	Prompt      string `yaml:"prompt"`
	Role        string `yaml:"role"`
	Plugin      string `yaml:"plugin"`
	Model       string `yaml:"model"`
	Temperature string `yaml:"temperature"`
}

const (
	configFileName = "config.yaml"
	version        = "0.1.5"
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
		Args:          cobra.ArbitraryArgs,
		RunE:          runCommand,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	tasksCmd = &cobra.Command{
		Use:   "tasks",
		Short: "LLM prompt chaining for complex tasks.",
		Long:  "Provide filepath to yaml file containing tasks to run.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				huh.NewFilePicker().
					Title("Select a yaml file containing the tasks to run:").
					AllowedTypes([]string{".yaml", "yml"}).
					Value(&appCfg.TasksPath).
					Picking(true).
					Height(10).
					Run()
			} else {
				appCfg.TasksPath = args[0]
			}

			handleTasks()
			return nil
		},
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
		WithTheme(huh.ThemeCharm()).
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
	rootCmd.AddCommand(tasksCmd)
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	flags := rootCmd.Flags()
	flags.StringVarP(&appCfg.Name, "plugin", "p", "openai", "The name of the plugin to use")
	flags.BoolVarP(&appCfg.ChoosePlugin, "choose-plugin", "P", false, "Choose the plugin to use")
	flags.StringVarP(&appCfg.Model, "model", "m", "", "The name of the model to use")
	flags.BoolVarP(&appCfg.ChooseAIModel, "choose-model", "M", false, "Choose the model to use")
	flags.BoolVarP(&appCfg.ChooseAIModel, "choose-model(deprecated)", "c", false, "Choose the model to use")
	flags.MarkHidden("choose-model(deprecated)")
	flags.MarkShorthandDeprecated("choose-model(deprecated)", "use -M instead")
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
			Title("What would you like to ask or discuss?").
			Value(&prompt).
			WithTheme(huh.ThemeCharm()).
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

func generateResponseForTasks(tasks Tasks) (string, error) {
	var out string

	for _, task := range tasks.Tasks {
		pluginCfg, err := getPluginConfig(task.Plugin, configPath)
		if err != nil {
			return "", err
		}
		if task.Temperature != "" {
			pluginCfg.Temperature = task.Temperature
		}

		pluginCfg.Role = task.Role
		prompt := out + task.Prompt

		res, err := pluginCfg.generateResponse(prompt, appCfg.Raw)
		if err != nil {
			return "", err
		}

		out = res
	}

	return out, nil
}

func handleTasks() error {
	tasksCfg, err := os.ReadFile(appCfg.TasksPath)
	if err != nil {
		return err
	}

	var tasks Tasks
	err = yaml.Unmarshal(tasksCfg, &tasks)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	var res string
	action := func() {
		res, err = generateResponseForTasks(tasks)
		cancel()
	}

	go action()
	_ = spinner.New().Title("Generating...").TitleStyle(lipgloss.NewStyle().Faint(true)).Context(ctx).Run()

	if err != nil {
		return err
	}

	fmt.Print(res)
	return nil
}

func choosePlugin() (string, error) {
	pluginCfgs, err := getAvailablePlugins(configPath)
	if err != nil {
		return "", err
	}

	var opts []huh.Option[string]
	for _, plugin := range pluginCfgs.Plugins {
		opts = append(opts, huh.Option[string]{
			Key:   plugin.Name,
			Value: plugin.Name,
		})
	}

	var plugin string
	huh.NewSelect[string]().
		Title("Choose a plugin:").
		Options(opts...).
		Value(&plugin).
		WithTheme(huh.ThemeCharm()).
		Run()

	return plugin, nil
}

func runCommand(cmd *cobra.Command, args []string) error {
	if appCfg.Version {
		fmt.Println(appName + " " + version)
		return nil
	}

	if appCfg.ChoosePlugin {
		pluginName, err := choosePlugin()
		if err != nil {
			return err
		}
		appCfg.Name = pluginName
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
	_ = spinner.New().Title("Generating...").TitleStyle(lipgloss.NewStyle().Faint(true)).Context(ctx).Run()

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
