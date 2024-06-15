package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/bitfield/script"
	"github.com/expr-lang/expr"
	extism "github.com/extism/go-sdk"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

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
			EnableWasi: true,
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
