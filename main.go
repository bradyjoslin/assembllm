package main

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/bitfield/script"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/expr-lang/expr"
	extism "github.com/extism/go-sdk"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"gopkg.in/yaml.v3"
)

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

type Tasks struct {
	IterationValuesIn string `yaml:"iterator_script"`
	IterationValues   []interface{}
	Tasks             []Task `yaml:"tasks"`
}

type Task struct {
	Name        string `yaml:"name"`
	Prompt      string `yaml:"prompt"`
	Role        string `yaml:"role"`
	Plugin      string `yaml:"plugin"`
	Model       string `yaml:"model"`
	Temperature string `yaml:"temperature"`
	PreScript   string `yaml:"pre_script"`
	PostScript  string `yaml:"post_script"`
}

const (
	configFileName = "config.yaml"
	version        = "0.4.1"
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
		Use:        "tasks",
		Hidden:     true,
		Deprecated: "Use the root command with the --workflow flag instead.\n",
		Short:      "LLM prompt chaining for complex tasks.",
		Args:       cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				huh.NewFilePicker().
					Title("Select a yaml file containing the tasks to run:").
					AllowedTypes([]string{".yaml", "yml"}).
					Value(&appCfg.WorkflowPath).
					Picking(true).
					Height(10).
					Run()
			} else {
				appCfg.WorkflowPath = args[0]
			}

			var prompt string
			if len(args) == 2 {
				prompt = args[1]
			}

			handleTasks(prompt)
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
	} else {
		// Read the existing config file
		configData, err := os.ReadFile(configPath)
		if err != nil {
			log.Fatalf("Unable to read config file: %v", err)
		}

		configDataUpdates := updatePlugins(configData)

		// Write the updated config back to the file
		err = os.WriteFile(configPath, configDataUpdates, 0644)
		if err != nil {
			log.Fatalf("Unable to write updated config to file: %v", err)
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
	flags.StringVarP(&appCfg.WorkflowPath, "workflow", "w", "", "The path to a workflow file")
	flags.BoolVarP(&appCfg.ChooseWorkflow, "choose-workflow", "W", false, "Choose a workflow to run")
	flags.BoolVarP(&appCfg.IteratorPrompt, "iterator", "i", false, "String array of prompts ['prompt1', 'prompt2']")
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

func httpGet(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func appendFile(content string, path string) (int64, error) {
	b, err := script.Echo(content).AppendFile(path)
	if err != nil {
		return 0, err
	}
	return b, nil
}

func readfile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func callExtismPlugin(source string, function string, input string) (string, error) {
	var wasm extism.Wasm

	if strings.HasPrefix(source, "https://") {
		wasm = extism.WasmUrl{
			Url: source,
		}
	} else {
		if !isFilePath(source) {
			return "", fmt.Errorf("file not found: %s", source)
		}

		wasm = extism.WasmFile{
			Path: source,
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
			EnableWasi: false,
		},
		[]extism.HostFunction{},
	)
	if err != nil {
		return "", err
	}
	if plugin == nil {
		return "", fmt.Errorf("plugin is nil")
	}

	_, out, err := plugin.Call(function, []byte(input))
	if err != nil {
		return "", err

	}
	response := string(out)

	return response, nil
}

func resend(to string, from string, subject string, body string) error {
	var html bytes.Buffer

	gm := goldmark.New(
		goldmark.WithExtensions(
			extension.Linkify,
			extension.Strikethrough,
			extension.Table,
		),
	)
	_ = gm.Convert([]byte(body), &html)

	escapedHTML := strconv.QuoteToASCII(html.String())
	escapedHTML = escapedHTML[1 : len(escapedHTML)-1] // Remove the extra double quotes

	payload := []byte(fmt.Sprintf(`{
	        "from": "%s",
	        "to": "%s",
	        "subject": "%s",
	        "html": "%s"
	    }`, from, to, subject, escapedHTML))

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(payload))

	if err != nil {
		return err
	}

	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return errors.New("RESEND_API_KEY is not set")
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("error sending email \n status code: %d\n%s", res.StatusCode, string(bodyBytes))
	}

	return nil
}

func workflowChain(path string, p string) (string, error) {
	var absPath string
	workflowDir := filepath.Dir(appCfg.WorkflowPath)

	// Join workflowDir with the provided path
	joinedPath := filepath.Join(workflowDir, path)

	// Convert joinedPath to an absolute path
	absPath, err := filepath.Abs(joinedPath)
	if err != nil {
		return "", fmt.Errorf("error loading workflow, check filepath: %v", err)
	}

	// Use absPath
	res, err := exec.Command("assembllm", "--raw", "-w", absPath, p).Output()
	if err != nil {
		return "", fmt.Errorf("error loading workflow, check filepath: %v", err)
	}
	return string(res), nil
}

func runExpr(input string, expression string) (string, error) {
	env := map[string]interface{}{
		"input":      input,
		"Get":        httpGet,
		"AppendFile": appendFile,
		"ReadFile":   readfile,
		"Extism":     callExtismPlugin,
		"Resend":     resend,
		"iterValue":  appCfg.CurrentIterationValue,
		"Workflow":   workflowChain,
	}

	program, err := expr.Compile(expression, expr.Env(env))
	if err != nil {
		return "", err
	}

	output, err := expr.Run(program, env)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", output), nil
}

func generateResponseForTasks(tasks Tasks) (string, error) {
	var out string

	for _, task := range tasks.Tasks {
		if task.PreScript != "" {
			s, err := runExpr(task.Prompt, task.PreScript)
			if err != nil {
				return "", err
			}
			task.Prompt = task.Prompt + s
		}

		var res string
		if task.Plugin != "" {
			pluginCfg, err := getPluginConfig(task.Plugin, configPath)
			if err != nil {
				return "", err
			}
			if task.Temperature != "" {
				pluginCfg.Temperature = task.Temperature
			}

			pluginCfg.Role = task.Role
			pluginCfg.Model = task.Model
			prompt := out + task.Prompt

			res, err = pluginCfg.generateResponse(prompt, true)
			if err != nil {
				return "", err
			}
		}
		if task.PostScript != "" {
			s, err := runExpr(res, task.PostScript)
			if err != nil {
				return "", err
			}
			res = s
		}

		out = res
	}

	if !appCfg.Raw {
		return glamour.Render(out, "dark")
	}
	return out, nil
}

func handleTasks(prompt string) error {
	tasksCfg, err := os.ReadFile(appCfg.WorkflowPath)
	if err != nil {
		return err
	}

	var tasks Tasks
	err = yaml.Unmarshal(tasksCfg, &tasks)
	if err != nil {
		return err
	}

	if tasks.IterationValuesIn == "" {
		tasks.IterationValues = []interface{}{nil}
	} else {
		env := map[string]interface{}{
			"input":      prompt,
			"Get":        httpGet,
			"AppendFile": appendFile,
			"ReadFile":   readfile,
			"Extism":     callExtismPlugin,
			"Resend":     resend,
		}

		program, err := expr.Compile(tasks.IterationValuesIn, expr.Env(env), expr.AsKind(reflect.Slice))
		if err != nil {
			return err
		}

		output, err := expr.Run(program, env)
		if err != nil {
			return err
		}

		tasks.IterationValues = output.([]interface{})
	}

	for i := range tasks.IterationValues {
		appCfg.CurrentIterationValue = tasks.IterationValues[i]

		if len(tasks.Tasks) > 0 {
			if prompt != "" {
				tasks.Tasks[0].Prompt = prompt + " " + tasks.Tasks[0].Prompt
			}
		}

		var res string
		action := func(tasks Tasks) {
			res, err = generateResponseForTasks(tasks)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		_ = spinner.New().
			Title("Generating...").
			TitleStyle(lipgloss.NewStyle().Faint(true)).
			Action(func() { action(tasks) }).
			Run()

		fmt.Print(res)
	}
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

	if appCfg.ChooseWorkflow {
		huh.NewFilePicker().
			Title("Select a workflow file:").
			AllowedTypes([]string{".yaml", "yml"}).
			Value(&appCfg.WorkflowPath).
			Picking(true).
			Height(10).
			Run()
	}

	if appCfg.WorkflowPath != "" {
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
		return handleTasks(prompt)
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
	var res string
	action := func(prompt string) {
		res, err = pluginCfg.generateResponse(prompt, appCfg.Raw)
	}

	var prompt string
	if appCfg.IteratorPrompt {
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
		for _, p := range prompts {
			_ = spinner.New().
				Title("Generating...").
				TitleStyle(lipgloss.NewStyle().Faint(true)).
				Action(func() { action(p) }).
				Run()

			fmt.Println(res)
		}
	} else {
		prompt = generatePrompt(args, appCfg.Raw)

		_ = spinner.New().
			Title("Generating...").
			TitleStyle(lipgloss.NewStyle().Faint(true)).
			Action(func() { action(prompt) }).
			Run()
	}

	if err != nil {
		return err
	}

	if !appCfg.IteratorPrompt {
		fmt.Print(res)
	}
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
