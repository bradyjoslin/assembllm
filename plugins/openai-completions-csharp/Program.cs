using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Runtime.InteropServices;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;
using Extism;

namespace Plugin;

public class Model
{
    [JsonPropertyName("name")]
    public string Name { get; set; }

    [JsonPropertyName("aliases")]
    public string[] Aliases { get; set; }
}

public class Message
{
    [JsonPropertyName("role")]
    public string Role { get; set; }

    [JsonPropertyName("content")]
    public string Content { get; set; }
}

public class CompletionRequest
{
    public RequestBody Body { get; set; }
    public string ApiKey { get; set; }
    public string Url { get; set; }
}

public class RequestBody
{
    [JsonPropertyName("model")]
    public string Model { get; set; }

    [JsonPropertyName("temperature")]
    public double Temperature { get; set; }

    [JsonPropertyName("messages")]
    public Message[] Messages { get; set; }
}

public class Choice
{
    [JsonPropertyName("message")]
    public Message Message { get; set; }
}

public class CompletionsResponse
{
    [JsonPropertyName("choices")]
    public Choice[] Choices { get; set; }
}

public class Program
{
    static readonly Model[] MODELS =
    [
        new Model { Name = "gpt-4o", Aliases = ["4o"] },
        new Model { Name = "gpt-4", Aliases = ["4"] },
        new Model { Name = "gpt-4-1106-preview", Aliases = ["128k"] },
        new Model { Name = "gpt-4-32k", Aliases = ["32k"] },
        new Model { Name = "gpt-3.5-turbo", Aliases = ["35t"] },
        new Model { Name = "gpt-3.5-turbo-1106", Aliases = ["35t-1106"] },
        new Model { Name = "gpt-3.5-turbo-16k", Aliases = ["35t16k"] },
        new Model { Name = "gpt-3.5", Aliases = ["35"] },
    ];

    public static void Main()
    {
        // Note: a `Main` method is required for the app to compile
    }

    [UnmanagedCallersOnly(EntryPoint = "models")]
    public static int Models()
    {
        var models = MODELS;
        Pdk.SetOutputJson<Model[]>(models, SourceGenerationContext.Default.ModelArray);

        return 0;
    }

    [UnmanagedCallersOnly(EntryPoint = "completion")]
    public static int Completion()
    {
        string prompt = Pdk.GetInputString();
        var (model, model_error) = GetModel();
        if (model_error != null)
        {
            Pdk.Log(LogLevel.Error, $"Error getting model: {model_error.Message}");
            Pdk.SetError(model_error.Message);
            return 1;
        }

        var (temperature, temperature_error) = GetTemperature();
        if (temperature_error != null)
        {
            Pdk.Log(LogLevel.Error, $"Error getting temperature: {temperature_error.Message}");
            Pdk.SetError(temperature_error.Message);
            return 1;
        }

        Pdk.TryGetConfig("role", out string role);

        Pdk.TryGetConfig("api_key", out string apiKey);
        if (apiKey == "")
        {
            Pdk.Log(LogLevel.Error, $"API key is required");
            Pdk.SetError("API key is required");
            return 1;
        }

        CompletionRequest completionRequest = new CompletionRequest
        {
            Body = new RequestBody
            {
                Model = model,
                Temperature = temperature,
                Messages =
        [
            new Message { Role = "system", Content = role },
            new Message { Role = "user", Content = prompt }
        ]
            },
            ApiKey = apiKey,
            Url = "https://api.openai.com/v1/chat/completions"
        };

        var (completionResponse, err3) = GetCompletionsResponse(completionRequest);
        if (err3 is not null)
        {
            Pdk.Log(LogLevel.Error, $"Error getting response: {err3.Message}");
            return 1;
        }

        Pdk.SetOutput(completionResponse.Choices[0].Message.Content + "\n");
        return 0;
    }

    public static (CompletionsResponse, Exception) GetCompletionsResponse(CompletionRequest cReq)
    {
        var request = new HttpRequest(cReq.Url)
        {
            Method = HttpMethod.POST
        };
        request.Headers.Add("Content-Type", "application/json");
        request.Headers.Add("Authorization", $"Bearer {cReq.ApiKey}");
        request.Body = Encoding.UTF8.GetBytes(JsonSerializer.Serialize(cReq.Body, SourceGenerationContext.Default.RequestBody));

        using (StreamReader reader = new StreamReader(new MemoryStream(request.Body), Encoding.UTF8))
        {
            string requestBody = reader.ReadToEnd();
            Pdk.Log(LogLevel.Info, requestBody);
        }
        var response = Pdk.SendRequest(request);
        if (response.Status != 200)
        {
            var responseBody = Encoding.UTF8.GetString(response.Body.ReadBytes());
            Pdk.Log(LogLevel.Info, $"Response body: {responseBody}");
            return (new CompletionsResponse(), new FormatException($"Error sending request: {response.Status}"));
        }

        try
        {
            var responseString = Encoding.UTF8.GetString(response.Body.ReadBytes());
            var completionsResponse = JsonSerializer.Deserialize(responseString, SourceGenerationContext.Default.CompletionsResponse);
            return (completionsResponse, null);
        }
        catch (JsonException ex)
        {
            Pdk.Log(LogLevel.Error, "Error unmarshalling response: " + ex.Message);
            return (null, new FormatException($"Error unmarshalling response: {ex.Message}", ex));
        }
    }

    public static (double, Exception) GetTemperature()
    {
        if (!Pdk.TryGetConfig("temperature", out string temperature))
        {
            Pdk.Log(LogLevel.Info, "Temperature not set, using default value");
            temperature = "0.7";
        }

        if (!double.TryParse(temperature, out double temperatureFloat))
        {
            Pdk.Log(LogLevel.Info, "No temp provided, setting default");
            return (0, new FormatException("Temperature must be a float"));
        }
        if (temperatureFloat < 0.0 || temperatureFloat > 1.0)
        {
            return (0, new ArgumentOutOfRangeException("Temperature must be between 0.0 and 1.0"));
        }

        return (temperatureFloat, null);
    }


    public static (string, Exception) GetModel()
    {
        if (!Pdk.TryGetConfig("model", out string model))
        {
            Pdk.Log(LogLevel.Info, "Model not set, using default value");
            return (MODELS[0].Name, null);
        }

        string validModel = null;
        foreach (var m in MODELS)
        {
            if (model == m.Name)
            {
                validModel = model;
                break;
            }
            if (m.Aliases.Any(alias => model == alias))
            {
                validModel = m.Name;
                break;
            }
        }
        if (string.IsNullOrEmpty(validModel))
        {
            return (null, new ArgumentException("Invalid model"));
        }

        return (validModel, null);
    }

}

[JsonSerializable(typeof(Model))]
[JsonSerializable(typeof(Model[]))]
[JsonSerializable(typeof(CompletionsResponse))]
[JsonSerializable(typeof(CompletionRequest))]
[JsonSerializable(typeof(RequestBody))]
[JsonSerializable(typeof(Message))]
[JsonSerializable(typeof(Message[]))]
[JsonSerializable(typeof(Choice))]
[JsonSerializable(typeof(Choice[]))]

public partial class SourceGenerationContext : JsonSerializerContext { }
