# Assembllm Plugins

## LLM plugins:

| Service    | Language   | Source                                                                            |
| ---------- | ---------- | --------------------------------------------------------------------------------- |
| OpenAI     | Rust       | [assembllm-openai](https://github.com/bradyjoslin/assembllm-openai)               |
| Perplexity | Rust       | [assembllm-perplexity](https://github.com/bradyjoslin/assembllm-perplexity)       |
| Cloudflare | Rust       | [assembllm-cloudflare](https://github.com/bradyjoslin/assembllm-cloudflare)       |
| Anthropic  | Go         | [assembllm-anthropic-go](https://github.com/bradyjoslin/assembllm-anthropic-go)   |
| OpenAI     | Go         | [assembllm-openai-go](https://github.com/bradyjoslin/assembllm-openai-go)         |
| OpenAI     | TypeScript | [assembllm-openai-ts](https://github.com/bradyjoslin/assembllm-openai-ts)         |
| OpenAI     | CSharp     | [assembllm-openai-csharp](https://github.com/bradyjoslin/assembllm-openai-csharp) |

## Script Plug-ins

### Assembllm HTML Tools

Source: https://github.com/bradyjoslin/assembllm-htmltools

### HTML Scraper

**Input**

The `scraper` function expects a JSON input with the following structure:

- `html`: The HTML content as a string.
- `selector`: A CSS selector to identify the elements to extract text from.

**Output**

The function outputs the text content of the matched elements.


### HTML Rewriter

The `htmlrewrite` function expects a JSON input with the following structure:

```json
{
  "html": "<html-content>",
  "rules": [
    {
      "selector": "<css-selector>",
      "html_content": "<new-html-content>"
    },
  ]
}
```

**Output**

The function outputs the modified HTML content as a string.
