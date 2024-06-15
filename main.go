package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	extism "github.com/extism/go-sdk"
	"github.com/spf13/cobra"
)

type App struct {
	Config  AppConfig
	RootCmd *cobra.Command
}

type AppConfig struct {
	Name                  string
	Model                 string
	ChooseAIModel         bool
	ChoosePlugin          bool
	ChooseWorkflow        bool
	Temperature           string
	Role                  string
	Raw                   bool
	Version               bool
	WorkflowPath          string
	IteratorPrompt        bool
	CurrentIterationValue interface{}
}

const (
	version = "0.5.0"
)

var (
	appCfg   AppConfig
	logLevel = extism.LogLevelOff
	appName  = filepath.Base(os.Args[0])
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

// Initializes the flags for the root command
func initializeFlags(app *App) {
	app.RootCmd.CompletionOptions.HiddenDefaultCmd = true
	app.RootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	flags := app.RootCmd.Flags()
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
	flags.StringVarP(&appCfg.WorkflowPath, "workflow", "w", "", "The path to a workflow file")
	flags.BoolVarP(&appCfg.ChooseWorkflow, "choose-workflow", "W", false, "Choose a workflow to run")
	flags.BoolVarP(&appCfg.IteratorPrompt, "iterator", "i", false, "String array of prompts ['prompt1', 'prompt2']")
	flags.SortFlags = false
}

// Generates a prompt for the chat completions
// If there is input from stdin, it will be included in the prompt
// If the user specified a prompt as an argument, it will be included in the prompt
// If there is no prompt, the user will be prompted to enter one
func generatePrompt(args []string, ask bool) string {
	var prompt string
	stdInStats, err := os.Stdin.Stat()
	if err != nil {
		fmt.Println("error getting stdin stats:", err)
		os.Exit(1)
	}

	if (stdInStats.Mode() & os.ModeCharDevice) == 0 {
		reader := bufio.NewReader(os.Stdin)
		s, err := io.ReadAll(reader)
		if err != nil {
			fmt.Println("error reading from stdin:", err)
			os.Exit(1)
		}
		prompt += string(s)
	}

	if len(args) == 1 {
		prompt += args[0]
	}

	if prompt == "" && ask {
		err := huh.NewInput().
			Title("What would you like to ask or discuss?").
			Value(&prompt).
			WithTheme(huh.ThemeCharm()).
			Run()
		if err != nil {
			fmt.Println("error getting input:", err)
			os.Exit(1)
		}
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

func choosePlugin() (string, error) {
	pluginCfgs, err := getAvailablePlugins(getConfigPath())
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

func createSpinner(action func()) error {
	return spinner.New().
		Title("Generating...").
		TitleStyle(lipgloss.NewStyle().Faint(true)).
		Action(action).
		Run()
}

func printVersion() {
	fmt.Println(appName + " " + version)
}

func chooseWorkflow() (string, error) {
	var workflowPath string
	err := huh.NewFilePicker().
		Title("Select a workflow file:").
		AllowedTypes([]string{".yaml", "yml"}).
		Value(&workflowPath).
		Picking(true).
		Height(10).
		Run()
	if err != nil {
		return "", fmt.Errorf("error choosing workflow: %v", err)
	}

	return workflowPath, nil
}

func buildIteratorPrompts(args []string) []string {
	var prompt string
	if len(args) > 0 {
		prompt = args[0]
	} else {
		huh.NewInput().
			Title("Enter the prompts in brackets separated by commas:").
			Value(&prompt).
			Placeholder("[prompt1, prompt2]").
			WithTheme(huh.ThemeCharm()).
			Run()
	}
	var prompts []string
	ps := strings.Trim(prompt, "[]")
	prompts = strings.Split(ps, ",")
	return prompts
}

func executeCompletion(pc CompletionPluginConfig, prompt string, spin bool) (string, error) {
	var res string
	var err error

	if spin {
		err = createSpinner(
			func() {
				res, err = pc.generateResponse(prompt, appCfg.Raw)
			},
		)
	} else {
		res, err = pc.generateResponse(prompt, appCfg.Raw)
	}
	if err != nil {
		return "", err
	}

	return res, nil
}

func runCommand(cmd *cobra.Command, args []string) error {
	if appCfg.Version {
		printVersion()
		return nil
	}

	if appCfg.ChooseWorkflow {
		wp, err := chooseWorkflow()

		appCfg.WorkflowPath = wp
		if err != nil {
			return err
		}
	}

	if appCfg.WorkflowPath != "" {
		return executeWorkflow(args)
	}

	if appCfg.ChoosePlugin {
		pluginName, err := choosePlugin()
		if err != nil {
			return err
		}
		appCfg.Name = pluginName
	}

	pluginCfg, err := getPluginConfig(appCfg.Name, getConfigPath())
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

	if appCfg.IteratorPrompt {
		prompts := buildIteratorPrompts(args)

		for _, p := range prompts {
			res, err := executeCompletion(pluginCfg, p, true)
			if err != nil {
				return err
			}

			fmt.Println(res)
		}
		return nil
	}

	prompt := generatePrompt(args, true)
	res, err := executeCompletion(pluginCfg, prompt, true)
	if err != nil {
		return err
	}

	fmt.Print(res)
	return nil
}

func main() {
	app := &App{
		Config: AppConfig{},
		RootCmd: &cobra.Command{
			Use:           appName + " [prompt]",
			Short:         "A WASM plug-in based CLI for AI chat completions",
			Args:          cobra.ArbitraryArgs,
			RunE:          runCommand,
			SilenceUsage:  true,
			SilenceErrors: true,
		},
	}

	initializeFlags(app)
	setupConfig()

	if err := app.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
