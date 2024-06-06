# assembllm

`assembllm` brings the power of LLM AIs to the command line with an extensible, WebAssembly-based plugin architecture.

- **LLM Chat Completions**: Supports building prompts piped from stdin and/or provided as an input argument.
- **Multi-AI Support**: Comes with built-in support for [OpenAI](https://platform.openai.com/docs/guides/text-generation/chat-completions-api), [Perplexity](https://docs.perplexity.ai/), and [Cloudflare AI](https://developers.cloudflare.com/workers-ai/models/#text-generation).
- **LLM Agent Chaining**: pass the output of one LLM as input to another, in a sequence or "pipeline", to perform complex tasks.
- **Plug-in Architecture**: Easily extend support for other LLMs. Plug-ins can be added via configuration without the need to recompile `assembllm`.
- **Cross-language support**: Create custom plugins in a variety of languages, including JavaScript, Rust, Go, C#, F#, AssemblyScript, Haskell, Zig, and C.

## Installing

```bash
# install with brew
brew tap bradyjoslin/assembllm
brew install bradyjoslin/assembllm/assembllm

# install with Go
go install github.com/bradyjoslin/assembllm
```

## Usage

```txt
$ assembllm -h
A WASM plug-in based CLI for AI chat completions

Usage:
  assembllm [prompt] [flags]
  assembllm [command]

Available Commands:
  tasks       LLM prompt chaining for complex tasks.

Flags:
  -p, --plugin string        The name of the plugin to use (default "openai")
  -P, --choose-plugin        Choose the plugin to use
  -m, --model string         The name of the model to use
  -M, --choose-model         Choose the model to use
  -t, --temperature string   The temperature to use
  -r, --role string          The role to use
      --raw                  Raw output without formatting
  -v, --version              Print the version
  -h, --help                 help for assembllm

Use "assembllm [command] --help" for more information about a command.
```

Quickly get completion responses using default plug-in and model:

![Demo](./assets/demo.gif)

Select from a list of models supported by each plug-in:

![Select Model Demo](./assets/choose_model_demo.gif)

Build complex prompts by piping from stdin:

![Curl Demo](./assets/piping_curl_demo.gif)

## LLM Chaining

Combine multiple LLMs using the `tasks` command, where the the results of each task feeds into the next. Let's walk through a sample task configuration.  At minimum each task needs a `name`, `plugin`, and `prompt`.  Here we use perplexity to generate the initial ideas for our topic.  Then conduct research and analysis on that output by first augmenting the results with information about Extism's GitHub repos using a `pre_script`, which calls a REST API and transforms the results to a more consice set of JSON, reducing LLM token usage.  Then we compose a blog post based on the research output and write the blog post to a local file using a `post_script`.  Finally, we read the blog post from the local file and generates a summary to stdout.

```yaml
tasks:
  - name: topic
    plugin: perplexity
    prompt: "ten bullets summarizing extism plug-in systems with wasm"
  - name: researcher
    plugin: openai
    pre_script: >
      (Get('https://api.github.com/orgs/extism/repos') | fromJSON()) 
      | map([
          {
            'name': .name, 
            'description': .description, 
            'stars': .stargazers_count, 
            'url': .html_url
          }
        ]) 
      | toJSON()
    role: "you are a technical research assistant"
    prompt: "analyze these capabilities against the broader backdrop of webassembly."
  - name: writer
    plugin: openai
    role: "you are a technical writer"
    prompt: "write a blog post on the provided research, avoid bullets, use prose and include section headers"
    temperature: 0.5
    model: 4o
    post_script: |
      AppendFile(input, "research_example_output.md")
  - name: reader
    plugin: openai
    pre_script: |
      ReadFile("research_example_output.md")
    role: "you are a technical reader"
    prompt: "summarize the blog post in 5 bullets"
```

Then run this task using `assebmllm`:

```sh
 assembllm tasks research_example_task.yaml
```

The results procuded this detailed blog post to [research_example_output.md](https://github.com/bradyjoslin/assembllm/blob/main/llm_chaining/research_example_output.md).

And printed this concise summary to stdout:

```md
  1. Language Agnosticism and Flexibility: Extism supports multiple
  programming languages through various Plug-in Development Kits (PDKs),
  enabling developers to use their preferred languages and existing codebases,
  consistent with the language-agnostic goals of the WASM ecosystem.

  2. Security and Sandboxing: Extism ensures the secure execution of untrusted
  code by leveraging WebAssembly's sandboxing and memory protection features,
  providing an additional layer of security for host applications.

  3. Host Functions and Extensibility: Extism allows plug-ins to import
  functions from the host application, facilitating powerful integrations such
  as database access and API usage, enhancing the functionality and
  flexibility of software.

  4. Use Cases and Practical Applications: Extism's versatility is showcased
  in various projects, including Function-as-a-Service (FaaS) platforms and web
  applications, aligning well with broader WASM trends in cloud computing,
  edge computing, IoT, and browser-based applications.

  5. Component Model and Future Roadmap: Extism is committed to evolving with
  the WASM ecosystem, actively tracking and planning the implementation of the
  Component Model to improve module interoperability and ease of use, ensuring
  it remains a cutting-edge tool for developers.
```

### Pre- and Post-Scripts

Create `pre_script` and `post_script` expressions with [Expr](https://expr-lang.org/), a Go-centric expression language designed to deliver dynamic configurations.  See the full language definition [here](https://expr-lang.org/docs/language-definition).  All expressions result in a single value.  

In addition to all of the functionality provided by Expr, `assebmllm` provides these additional functions that you can use in your expressions:

- **Get**: perform http Get calls within functions
- **ReadFile**: read files from your local filesystem
- **AppendFile**: appends content to file, creating if it doesn't exist

In addition to these functions an `input` variable is provided with the contents of the prompt at that stage of the chain.

A `pre_script` is run before sending the prompt to the LLM.  The output of a `pre_script` is appended to the prompt at that stage of the chain.

A `post_script` is run after sending the prompt to the LLM and the `input` value availabe in the expression is the LLM results.  Unlike a `pre_script`, `post_script` the expression's output *replaces* instead of appends to the prompt at that stage of the chain.  If you would like to pass the prompt along from a `post_script`, you must do so explicitly.  For example, if you'd like to write the current prompt to a file and also pass it to the next LLM: 

```yml
...
  - name: file_writer
    post_script: |
      let b = AppendFile(input, "research_example_output.md");
      input
...
```

### Chaining with Bash Scripts

Alternatively, you can chain LLM responses using bash scripts:

```sh
#!/bin/bash

TOPIC="ten bullets summarizing extism plug-in systems with wasm"
RESEARCHER="you are a technical research assistant"
ANALYSIS="analyze these capabilities against the broader backdrop of webassembly."
WRITER="you are a technical writer specializing in trends, skilled at helping developers understand the practical use of new technology described from first principles"
BLOG_POST="write a blog post on the provided research, avoid bullets, use prose and include section headers"

assembllm -p perplexity "$TOPIC" \
| assembllm -r "$RESEARCHER" "$ANALYSIS" \
| assembllm --raw -r "$WRITER" "$BLOG_POST" \
| tee research_example_output.md
```

Or grab a pre-built binary from [releases](https://github.com/bradyjoslin/assembllm/releases).

## Plugins

Plug-ins are powered by [Extism](https://extism.org), a cross-language framework for building web-assembly based plug-in systems.  `assembllm` acts as a [host application](https://extism.org/docs/concepts/host-sdk) that uses the Extism SDK to and is responsible for handling the user experience and interacting with the LLM chat completion plug-ins which use Extism's [Plug-in Development Kits (PDKs)](https://extism.org/docs/concepts/pdk).

### Sample Plugins

Sample plugins are provided in the `/plugins` directory implemented using Rust, TypeScript, Go, and C#.   These samples are implemented in the default configuration on install.

### Plug-in Configuration

`assembllm` chat completion plugins are defined in `config.yaml` that is stored in `~/.assembllm`.  The first plug-in in the configuration file will be used as the default.

The provided plug-in configuration is used to define an [Extism manifest](https://extism.org/docs/concepts/manifest/) that `assembllm` uses to load the Wasm module, grant it the relevant permissions, and provide configuration data.  Wasm is sandboxed by default, unable to access the filesystem, make network calls, or access system information like environment variables unless explicitly granted by the host.

Let's walk through a sample configuration as defined below. We're importing a plug-in named `openai` whose Wasm source is loaded from a remote URL.  A hash is provided to confirm the integrity of the Wasm source. The `apiKey` for the plug-in will be loaded from an environment variable named `OPENAI_API_KEY` and passed as a configuration value to the plug-in.  The base URL the plug-in will use to make API calls to the OpenAI API is provided, granting the plug-in permission to call that resource as an allowed host.  Lastly, we set a default model, which is passed as a configuration value to the plug-in.  

```yml
completion-plugins:
  - name: openai
    source: https://cdn.modsurfer.dylibso.com/api/v1/module/114e1e892c43baefb4d50cc8b0e9f66df2b2e3177de9293ffdd83898c77e04c7.wasm
    hash: 114e1e892c43baefb4d50cc8b0e9f66df2b2e3177de9293ffdd83898c77e04c7
    apiKey: OPENAI_API_KEY
    url: api.openai.com
    model: 4o
...
```

Here is the full list of available plug-in configuration values:

- `name`: unique name for the plugin.
- `source`: wasm file location, can be a file path or http location.
- `hash`: sha 256-based hash of the wasm file for validation.  Optional, but recommended.
- `apiKey`: environment variable name containing the API Key for the service the plug-in uses
- `accountId`: environment variable name containing the AccountID for the plugin's service.  Optional, used by some services like [Cloudflare](https://developers.cloudflare.com/workers-ai/get-started/rest-api/#1-get-api-token-and-account-id).
- `url`: the base url for the service used by the plug-in. 
- `model`: default model to use.
- `wasi`: whether or not the plugin requires WASI.

### Plug-in Architecture

To be compatible with `assembllm`, each plugin must expose two functions via the PDK:

- **Models**: provides a list of models supported by the plug-in
- **Completion**: takes a string prompt and returns a completions response

### models Function

A `models` function should be exported by the plug-in and return an array of models supported by the LLM. Each object has the following properties:

- `name` (string): The name of the model.
- `aliases` (array): An array of strings, each string is an alias for the model.

Sample response:

```json
[
  {
    "name": "gpt-4o",
    "aliases": ["4o"]
  },
  {
    "name": "gpt-4",
    "aliases": ["4"]
  },
  {
    "name": "gpt-3.5",
    "aliases": ["35"]
  }
]
```

Here's a JSON Schema for the objects in the `models` array:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "array",
  "items": {
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
      }
    },
    "required": ["name", "aliases"]
  }
}
```

### completion Function

A `completion` function should be exported by the plug-in that takes the prompt and configuration as input and provides the chat completion response as output.

The plug-in is also provided configuration data from the `assembllm` host:

- `api_key`: user's API Key to use for the API service call
- `accountId`: Account ID for the plug-in service. Used by some services like [Cloudflare](https://developers.cloudflare.com/workers-ai/get-started/rest-api/#1-get-api-token-and-account-id).
- `model`: LLM model to use for completions response
- `temperature`: temperature value for the completion response
- `role`: prompt to use as the system message for the prompt
