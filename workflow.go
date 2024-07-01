package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/expr-lang/expr"
	"gopkg.in/yaml.v3"
)

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
	Tools       []Tool `yaml:"tools,omitempty"`
}

func getAbsolutePath(path string) (string, error) {
	workflowDir := filepath.Dir(appCfg.WorkflowPath)
	joinedPath := filepath.Join(workflowDir, path)
	return filepath.Abs(joinedPath)
}

func workflowChain(path string, p string) (string, error) {
	absPath, err := getAbsolutePath(path)
	if err != nil {
		return "", fmt.Errorf("error loading workflow, check filepath: %v", err)
	}

	res, err := exec.Command("assembllm", "--raw", "-w", absPath, p).Output()
	if err != nil {
		return "", fmt.Errorf("error loading workflow: %v\n%v\n%v", absPath, string(res), err)
	}
	return string(res), nil
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
			pluginCfg, err := getPluginConfig(task.Plugin, getConfigPath())
			if err != nil {
				return "", err
			}
			if task.Temperature != "" {
				pluginCfg.Temperature = task.Temperature
			}

			pluginCfg.Role = task.Role
			pluginCfg.Model = task.Model
			prompt := out + task.Prompt

			if task.Tools != nil {
				res, err = pluginCfg.generateResponseWithTools(prompt, task.Tools)
				if err != nil {
					return "", err
				}
			} else {

				res, err = pluginCfg.generateResponse(prompt, true)
				if err != nil {
					return "", err
				}
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

	var res string
	for i := range tasks.IterationValues {
		appCfg.CurrentIterationValue = tasks.IterationValues[i]

		if len(tasks.Tasks) > 0 {
			if prompt != "" {
				tasks.Tasks[0].Prompt = prompt + " " + tasks.Tasks[0].Prompt
			}
		}

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

	if appCfg.Feedback {
		var rerun bool
		huh.NewConfirm().Title("Would you like to provide feedback and rerun the workflow?").Value(&rerun).Run()

		if rerun {
			var feedback string
			huh.NewInput().Title("Provide your feedback or follow-up question:").Value(&feedback).Run()
			return handleTasks("you were prompted with " + prompt + "and responded with " + res + " the user provided this feedback: " + feedback)
		}
	}

	return nil
}

func executeWorkflow(args []string) error {
	prompt := generatePrompt(args, false)
	return handleTasks(prompt)
}
