# assembllm

<p><div align="center">
<img src="./assets/assembllm_banner.png" alt="Banner Image" style="width: 100%; max-width: 800px; height: auto;">
</div></p>

A versatile CLI tool designed to combine multiple Large Language Models (LLMs) using a flexible task-based system. With a unique WebAssembly-based plugin architecture, it supports seamless integration of various AI models and custom scripts.

### Key Features:

- **Multi-Model Support**: Integrates with [OpenAI](https://platform.openai.com/docs/guides/text-generation/chat-completions-api), [Perplexity](https://docs.perplexity.ai/), and [Cloudflare AI](https://developers.cloudflare.com/workers-ai/models/#text-generation), among others.
- **Plugin Architecture**: Extensible via WebAssembly plugins written in a variety of languages, including JavaScript, Rust, Go, and C#, using [Extism](https://extism.org/).
- **Task Chaining**: Automate workflows by chaining multiple tasks where outputs of one feed into the next.
- **Flexible Scripting**: Use pre- and post-scripts for data transformation and integration.

## Basic Usage

### Simple Commands

You can quickly utilize the power of LLMs with simple commands and integrate assembllm with other tools via bash pipes for enhanced functionality:

![Demo](./assets/basic_demo.gif)

## Advanced Prompting with Workflows

For more complex prompts, including the ability to create prompt pipelines, define and chain tasks together with workflows.  We have a [library of workflows](https://github.com/bradyjoslin/assembllm/tree/main/workflows) you can use as examples and templates, let's walk through one together here.

### Example Workflow Configuration

This example demonstrates how to chain multiple tasks together to generate, analyze, compose, and summarize content:

- **Generate Topic Ideas**: Use Perplexity to generate initial ideas.
- **Conduct Research**: Augment the generated ideas with data from Extism's GitHub repositories using a pre-script.
- **Write Blog Post**: Compose a blog post based on the research, write it to a file, and send as an email.
- **Summarize Blog Post**: Read the blog post from a file and generate a summary.

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
      let _ = AppendFile(input, "research_example_output.md");
      Resend("example@example.com", "info@notifications.example.com", "Extism Research", input)

  - name: reader
    plugin: openai
    pre_script: |
      ReadFile("research_example_output.md")
    role: "you are a technical reader"
    prompt: "summarize the blog post in 5 bullets"
```

Run this workflow with:

```sh
 assembllm --workflow research_example_task.yaml
```

After running the above workflow, you can expect as outputs a detailed blog post saved to a file and sent as an email, and a concise summary printed to stdout.

### Workflow Prompts

Workflows in `assembllm` can optionally take a prompt from either standard input (stdin) or as an argument. The provided input is integrated into the prompt defined in the first task of the workflow.

**Key Points**:

- **Optional Input**: You can run workflows without providing a prompt as input.
- **Optional First Task Prompt**: The prompt in the first task of a workflow is also optional.
- **Combining Prompts**: If both are provided, they are combined to form a unified prompt for the first task.

This flexibility allows workflows to be dynamic and adaptable based on user input.

### Pre-Scripts and Post-Scripts

assembllm allows the use of pre-scripts and post-scripts for data transformation and integration, providing flexibility in how data is handled before and after LLM processing. These scripts can utilize various functions to fetch, read, append, and transform data.

Expressions are powered by [Expr](https://expr-lang.org/), a Go-centric expression language designed to deliver dynamic configurations.  All expressions result in a single value. See the full language definition [here](https://expr-lang.org/docs/language-definition). 

In addition to all of the functionality provided by Expr, these functions are available in expressions:

- **Get**: perform http GET calls within functions
  - **Signature**: Get(url: str) -> str
  - **Parameters**: url (str): The URL to fetch data from.
  - **Returns**: Response data as a string

- **ReadFile**: read files from your local filesystem
  - **Signature**: ReadFile(filepath: str) -> str
  - **Parameters**: filepath (str): The path to the file to read.
  - **Returns**: File content as a string.

- **AppendFile**: appends content to file, creating if it doesn't exist
  - **Signature**: AppendFile(content: str, filepath: str) -> (int64, error)
  - **Parameters**:
    - content (str): The content to append.
    - filepath (str): The path to the file to append to.
  - **Returns**: Number of bytes written as int64 or an error

- **Resend**: sends content as email using [Resend](https://resend.com/)
  - **Signature**: Resend(to: str, from: str, subject: str, body: str) -> Error
  - **Parameters**:
    - to (str): Email to field
    - from (str): Email from field
    - subject (str): Email subject
    - body (str): Email body, automatically converted from markdown to HTML
  - **Returns**: Error, if one occured
  - **Requires**: [Resend API key](https://resend.com/docs/dashboard/api-keys/introduction) set to `RESEND_API_KEY` environment variable

- **Extism**: calls a wasm function, source can be a file or url
  - **Signature**: Extism(source: str, function_name: str, args: list) -> str
  - **Parameters**:
    - source (str): The source of the WebAssembly function (file or URL).
    - function_name (str): The name of the function to call.
    - args (list): A list of arguments to pass to the function.
  - **Returns**: Result of the WebAssembly function call as a string.

In addition to these functions, an `input` variable is provided with the contents of the prompt at that stage of the chain.

A `pre_script` is run before sending the prompt to the LLM.  The output of a `pre_script` is appended to the prompt at that stage of the chain.

A `post_script` is run after sending the prompt to the LLM, therefore the `input` value availabe is the LLM's response.  Unlike a `pre_script`, a `post_script`'s output *replaces* instead of appends to the prompt at that stage of the chain, so if you would like to pass the prompt along from a `post_script`, you must do so explicitly.  For example, if you'd like to write the current LLM results to a file and also pass those results to the next LLM: 

```yml
...
  - name: file_writer
    post_script: |
      let _ = AppendFile(input, "research_example_output.md");
      input
...
```

Here's an example of calling wasm using the `Extism` function within expressions:

```sh
tasks:
  - name: topic
    plugin: openai
    prompt: "tell me a joke"
    post_script: |
      let wasm = "https://github.com/extism/plugins/releases/latest/download/count_vowels.wasm";
      let vowelCount = Extism(wasm, "count_vowels", input);
      let vowels = (vowelCount | fromJSON()).count | string();
      input + "\n\n vowels: " + vowels
```

Example results:

```txt
  Sure, here's a light-hearted joke for you:

  Why don't skeletons fight each other?

  They don't have the guts.

  vowels: 29
```

### Chaining with Bash Scripts

While assembllm provides a powerful built-in workflow feature, you can also chain LLM responses directly within Bash scripts for simpler automation. Here’s an example:

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

**Explanation**:

- **TOPIC**: Generate initial topic ideas.
- **RESEARCHER**: Analyze the generated ideas.
- **ANALYSIS**: Provide a deeper understanding of the topic.
- **WRITER**: Compose a detailed blog post based on the research.

This script demonstrates how you can chain multiple LLM commands together, leveraging `assembllm` to process and transform data through each stage. This approach offers an alternative to the built-in workflow feature for those who prefer using Bash scripts.

## Plugins

Plug-ins are powered by [Extism](https://extism.org), a cross-language framework for building web-assembly based plug-in systems.  `assembllm` acts as a [host application](https://extism.org/docs/concepts/host-sdk) that uses the Extism SDK to and is responsible for handling the user experience and interacting with the LLM chat completion plug-ins which use Extism's [Plug-in Development Kits (PDKs)](https://extism.org/docs/concepts/pdk).

### Sample Plugins

Sample plugins are provided in the `/plugins` directory implemented using Rust, TypeScript, Go, and C#. These samples are implemented in the default configuration on install.

### Plug-in Configuration

Plugins are defined in `config.yaml`, stored in `~/.assembllm`. The first plugin in the configuration file will be used as the default.

The provided plugin configuration defines an [Extism manifest](https://extism.org/docs/concepts/manifest/) that `assembllm` uses to load the Wasm module, grant it relevant permissions, and provide configuration data. By default, Wasm is sandboxed, unable to access the filesystem, make network calls, or access system information like environment variables unless explicitly granted by the host.

Let's walk through a sample configuration. We're importing a plugin named openai whose Wasm source is loaded from a remote URL. A hash is provided to confirm the integrity of the Wasm source. The `apiKey` for the plugin will be loaded from an environment variable named `OPENAI_API_KEY` and passed as a configuration value to the plugin. The base URL the plugin will use to make API calls to the OpenAI API is provided, granting the plugin permission to call that resource as an allowed host. Lastly, we set a default model, which is passed as a configuration value to the plugin.

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
