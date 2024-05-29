# assembllm

`assembllm` brings LLM-based AI to the command line and provides an extensible plugin-based architecture. Write your own LLM plugins for `assembllm` in JavaScript, Rust, Go, C#, F#, AssemblyScript, Haskell, Zig, or C.

**Features:**

- **LLM Chat Completions**: supports building LLM prompts piped from stdin and provided as an input argument
- **Multi-AI Support**: Supports OpenAI and Perplexity out of the box
- **Plugin Architecture**: Easily extend support for other LLM by adding or creating plugins

## Usage

```bash
$ assembllm -h
Usage:
  assembllm [prompt] [flags]

Flags:
  -p, --plugin string        The name of the plugin to use (default "openai")
  -m, --model string         The name of the model to use
  -c, --choose-model         Choose the model to use
  -t, --temperature string   The temperature to use
  -r, --role string          The role to use
      --raw                  Raw output without formatting
  -h, --help                 help for ayeye
```

## Plugins

Plug-ins are powered by [Extism](https://extism.org), a cross-language framework for building web-assembly based plug-in systems.

### Plugin Configuration

`assembllm` chat completion plugins are defined in `config.yml`.  Each plugin is defined by:

- `name`: unique name for the plugin
- `source`: a reference to the built wasm file for the plug-in.  Can be defined by specifying a path or an http location
- `hash`: sha 256-based hash of the specified wasm file for validation.  Optional, but recommended.
- `apiKey`: the environment variable containing the API Key for the plugin's service
- `url`: the base url for the service used by the plug-in.  By default the plug-ins cannot make http calls, this grants access to the plug-in to call the API resource.
- `model`: default model to use if one isn't specified
- `wasi`: whether or not the plugin requires WASI

`assembllm` acts as a [host application](https://extism.org/docs/concepts/host-sdk) that uses the Extism SDK to and is responsible for handling the user experience and interacting with the LLM chat completion plug-ins defined using Extism's [Plug-in Development Kits (PDKs)](https://extism.org/docs/concepts/pdk).

To be compatible with `assembllm`, each plugin must expose two functions via the PDK:

- **Models**: provides a list of models supported by the plugin
- **Completion**: takes a string prompt and returns a completions response

Each plugin is also provided a configuration data:

- `api_key`: user's API Key to use for the API service call
- `model`: LLM model to use for completions response
- `temperature`: temperature value for the completion response
- `role`: prompt to use as the system message for the prompt

### models Function

The `models` function returns an array of objects. Each object has the following properties:

- `name` (string): The name of the model.
- `aliases` (array): An array of strings, each string is an alias for the model.
- `max_input_chars` (integer): The maximum number of input characters the model can handle.
- `fallback` (string): The name of the fallback model.

Sample response:

```json
[
  {
    "name": "gpt-4o",
    "aliases": [
      "4o"
    ],
    "max_input_chars": 128000,
    "fallback": "gpt-4"
  },
  {
    "name": "gpt-4",
    "aliases": [
      "4"
    ],
    "max_input_chars": 24500,
    "fallback": "gpt-3.5-turbo"
  },
  {
    "name": "gpt-3.5",
    "aliases": [
      "35"
    ],
    "max_input_chars": 12250,
    "fallback": ""
  }
]
```

Here's a JSON Schema for the objects in the `models` array:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "name": {
      "type": "string"
    },
    "aliases": {
      "type": "array",
      "items": {
        "type": "string"
      }
    },
    "max_input_chars": {
      "type": "integer"
    },
    "fallback": {
      "type": "string"
    }
  },
  "required": ["name", "aliases", "max_input_chars", "fallback"]
}
```

## Sample Plugins

Sample plugins are provided in the `/plugins` directory and show how to build plug-ins using C#, Rust, and Go.   These samples are also used in the default configuration.

## Installing

```bash
# install with GO
go install github.com/bradyjoslin/assembllm
```
