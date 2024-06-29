# Scrape and Summarize Workflow Example

**Overview**

[scrape_then_summarize.yaml](./scrape_then_summarize.yaml) uses workflow chaining and [Assembllm HTML Tools](https://github.com/bradyjoslin/assembllm-htmltools) to compose a workflow that scrapes a URL and then summarizes the content.

Takes a json array of objects with the following keys:
- **url**: The URL to scrape
- **selector**: The CSS selector to scrape

**Sample Output**:

![scrape summarize gif](scrape_then_summarize.gif)

**Usage**

Example:

```
assembllm -w scrape_then_summarize.yaml '[{"url": "https://bradyjoslin.com/blog/remote-vs-code/", "selector": "div.post-content"}]'
```

## Workflow

The entire workflow simply calls two workflows, one to scrape content from a web page, the other to provide the summary:

```yaml
# scrape_then_summarize.yaml
iterator_script: |
  [input]

tasks:
  - name: scrape
    post_script: |
      let content = Workflow("scrape.yaml", iterValue);
      Workflow("summarize.yaml", content)
```

## Dependent Workflows

For deeper intuition, let's dive into how the called `scrape.yaml` and `summarize.yaml` workflows operate.

### scrape.yaml

Gets the contents of a remote URL, then scrape the contents of a specified selector using HTML Tools, an Extism-based wasm plug-in.

```yaml
# scrape.yaml
iterator_script: |
  input | fromJSON()

tasks:
  - name: scrape
    post_script: |
      let wasm = "https://github.com/bradyjoslin/assembllm-htmltools/releases/latest/download/assembllm-htmltools.wasm";
      let content = Get(iterValue.url);
      let params = (
        {
          "html": content, 
          "selector": iterValue.selector
        } 
        | toJSON()
      );
      Extism(wasm, "scraper", params)
```

### summarize.yaml

Optimized for taking an article's content and providing an effective summary 

```yaml
# summarize.yaml
tasks:
  - plugin: anthropic
    prompt: |
      # Task Objective

      Give a detailed analysis of the topic supported content provided. Response must 
      be 250 words or less.

      ## Task Details

      Be insightful. Think about the author's claims deeply and express the strengths
      and weaknesses of their argument.  Where practical, think about how the topic
      can impact a person's daily life, or an industry long-term, or is it just
      passing information with little impact.
```