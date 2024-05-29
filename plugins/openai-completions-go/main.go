package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/extism/go-pdk"
)

type Model struct {
	Name          string   `json:"name"`
	Aliases       []string `json:"aliases"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CompletionRequest struct {
	Body   RequestBody
	ApiKey string
	Url    string
}

type RequestBody struct {
	Model       string    `json:"model"`
	Temperature float64   `json:"temperature"`
	Messages    []Message `json:"messages"`
}

type Choice struct {
	Message Message `json:"message"`
}

type CompletionsResponse struct {
	Choices []Choice `json:"choices"`
}

var models = []Model{
	{
		Name:          "gpt-4o",
		Aliases:       []string{"4o"},
	},
	{
		Name:          "gpt-4",
		Aliases:       []string{"4"},
	},
	{
		Name:          "gpt-4-1106-preview",
		Aliases:       []string{"128k"},
	},
	{
		Name:          "gpt-4-32k",
		Aliases:       []string{"32k"},
	},
	{
		Name:          "gpt-3.5-turbo",
		Aliases:       []string{"35t"},
	},
	{
		Name:          "gpt-3.5-turbo-1106",
		Aliases:       []string{"35t-1106"},
	},
	{
		Name:          "gpt-3.5-turbo-16k",
		Aliases:       []string{"35t16k"},
	},
	{
		Name:          "gpt-3.5",
		Aliases:       []string{"35"},
	},
}

//go:export models
func Models() int32 {
	modelsJson, err := json.Marshal(models)
	if err != nil {
		pdk.OutputString("Error converting models to JSON: " + err.Error())
		return 1
	}

	pdk.Log(pdk.LogInfo, "Returning models")
	pdk.OutputString(string(modelsJson))
	return 0
}

func getTemperature() (float64, error) {
	temperature, _ := pdk.GetConfig("temperature")
	if temperature == "" {
		pdk.Log(pdk.LogInfo, "Temperature not set, using default value")
		temperature = "0.7"
	}

	temperatureFloat, err := strconv.ParseFloat(temperature, 32)
	if err != nil {
		return 0, fmt.Errorf("Temperature must be a float: %v", err)
	}
	if temperatureFloat < 0.0 || temperatureFloat > 1.0 {
		return 0, fmt.Errorf("Temperature must be between 0.0 and 1.0")
	}

	return temperatureFloat, nil
}

func getModel() (string, error) {
	model, ok := pdk.GetConfig("model")
	if !ok {
		pdk.Log(pdk.LogInfo, "Model not set, using default value")
		return models[0].Name, nil
	}

	var validModel string
	for _, m := range models {
		if model == m.Name {
			validModel = model
			break
		}
		for _, alias := range m.Aliases {
			if model == alias {
				validModel = m.Name
				break
			}
		}
	}
	if validModel == "" {
		return "", fmt.Errorf("Invalid model")
	}

	return validModel, nil
}

func (cReq CompletionRequest) getCompletionsResponse() (CompletionsResponse, error) {
	jsonData, err := json.Marshal(cReq.Body)
	if err != nil {
		fmt.Println(err)
		return CompletionsResponse{}, err
	}

	req := pdk.NewHTTPRequest(pdk.MethodPost, cReq.Url)
	req.SetBody(jsonData)
	req.SetHeader("Content-Type", "application/json")
	req.SetHeader("Authorization", "Bearer "+cReq.ApiKey)

	res := req.Send()
	if res.Status() != 200 {
		pdk.Log(pdk.LogError, fmt.Sprintf("Error sending request: %v", res.Status()))
		return CompletionsResponse{}, fmt.Errorf("Error sending request: %v", res.Status())
	}

	body := res.Body()

	var completionsResponse CompletionsResponse
	err = json.Unmarshal([]byte(body), &completionsResponse)
	if err != nil {
		pdk.Log(pdk.LogError, "Error unmarshalling response: "+err.Error())
		return CompletionsResponse{}, fmt.Errorf("Error unmarshalling response: %v", err)
	}
	return completionsResponse, nil
}

//go:export completion
func Completion() int32 {
	prompt := pdk.InputString()

	api_key, ok := pdk.GetConfig("api_key")
	if !ok {
		pdk.Log(pdk.LogError, "Error getting api_key")
		return 1
	}

	role, ok := pdk.GetConfig("role")
	if !ok {
		pdk.Log(pdk.LogInfo, "Role not set")
	}

	temperature, err := getTemperature()
	if err != nil {
		pdk.Log(pdk.LogError, fmt.Sprintf("Error getting temperature: %v", err.Error()))
		return 1
	}
	pdk.Log(pdk.LogInfo, fmt.Sprintf("Temperature: %v", temperature))

	model, err := getModel()
	if err != nil {
		pdk.Log(pdk.LogError, fmt.Sprintf("Error getting model: %v", err.Error()))
		return 1
	}
	pdk.Log(pdk.LogInfo, fmt.Sprintf("Model: %v", model))

	pdk.Log(pdk.LogInfo, "Prompt: "+prompt)

	completionRequest := CompletionRequest{
		Body: RequestBody{
			Model:       model,
			Temperature: temperature,
			Messages: []Message{
				{
					Role:    "system",
					Content: role,
				},
				{
					Role:    "user",
					Content: prompt,
				},
			},
		},
		ApiKey: api_key,
		Url:    "https://api.openai.com/v1/chat/completions",
	}

	completionResponse, err := completionRequest.getCompletionsResponse()

	if err != nil {
		pdk.Log(pdk.LogError, fmt.Sprintf("Error getting completions response: %v", err.Error()))
		return 1
	}

	pdk.OutputString(completionResponse.Choices[0].Message.Content)
	return 0
}

func main() {}
