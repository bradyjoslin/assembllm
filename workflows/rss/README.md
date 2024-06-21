# RSS Workflow Example

**Overview**

[This workflow](./rss.yaml) retrieves an RSS feed provided as a prompt, extracts the titles and URLs of the first five stories, and then calls an external workflow to generate summaries for each story. Finally, a post script formats the output by adding the title from the RSS feed as a markdown header and prints the summary.

**Sample Output**:

![rss gif](rss.gif)

**Usage**

```sh
assembllm -w rss.yaml <prompt (RSS feed URL)>
```

Example:

```
assembllm -w ~/Projects/assembllm/workflows/rss/rss.yaml https://news.ycombinator.com/rss
```

## Step by Step Guide

### Step 1: Fetch and Parse RSS Feed

The provided RSS feed URL is fetched using the `Get` function.  The RSS feed is parsed, where the first five items in the feed are mapped into a list of strings that contain the post title and URL separated by `---`.

```yaml
iterator_script: |

  let rss = Get(input);

  let items = split(rss, "<item>")[1:6];

  items | map(
      join(
        split(#, "<title>") 
        | last()
        | split("</title>")
        | map(
          split(#, "<link>") 
          | last() + " --- "
          | split("</link>") 
          | first()
       )
     )
  )
```

**Sample Output:**

For the URL https://news.ycombinator.com/rss, the output would be a list like:

```txt
[
  "Agricultural drones are transforming rice farming in the Mekong River delta --- https://hakaimagazine.com/videos-visuals/rice-farming-gets-an-ai-upgrade/",
  "1\25-scale Cray C90 wristwatch --- http://www.chrisfenton.com/1-25-scale-cray-c90-wristwatch/",
  "How We Made the Deno Language Server Ten Times Faster --- https://deno.com/blog/optimizing-our-lsp",
  "Autonomous vehicles are great at driving straight --- https://spectrum.ieee.org/autonomous-vehicles-great-at-straights",
  "What happens to our breath when we type, tap, scroll --- https://www.npr.org/2024/06/10/1247296780/screen-apnea-why-screens-cause-shallow-breathing"
]
```

This list, generated within the iteration script, is then iterated over, with each value being processed through the series of tasks defined in the workflow.

### Step 2: Summarize Articles and Format Output

For each item in the list, an external workflow is called to generate an article summary. The title of each article is then formatted as a markdown header above the summary. Being able to call workflows from workflow allows for modularity and reuse.  In this case, we defer to an existing workflow that uses an online Perplexity model  that can get content from a webpage when given a URL in a prompt.

```yaml
tasks:
  - post_script: |
      let res = Workflow("../article_summarizer.yaml", iterValue);

      let title = split(iterValue, "---") | first();
      "# " + title + "\n" + res
```

Here's the definition of the summarization workflow:

```yaml
# article_summarizer.yaml
tasks:
  - plugin: perplexity
    prompt: |
      # Task Objective

      Give a detailed analysis of the topic supported by a URL source provided.
      Provide references, if available.  Include the URL provided at the end of
      the summary.  Response must be 250 words or less.

      ## Task Details

      Be insightful. Think about the author's claims deeply and express the strengths
      and weaknesses of their argument.  Where practical, think about how the topic
      can impact a person's daily life, or an industry long-term, or is it just
      passing information with little impact.
```

That's it, the full workflow can be found [here](./rss.yaml).
