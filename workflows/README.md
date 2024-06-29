# Example Workflows

These samples are intended to be useful on their own while also serving as cookbooks that can be used as references to build your own workflows.  Contributions welcome!

## End of Life

Provides end of life information on software products provided within a prompt.  Explains how to use LLM function calling to get structured data from an unstructured prompt and how to perform API calls in a workflow script. ([details](./eol/))

<img src="eol/eol.gif" width="800px">

## RSS

Takes a URL to an RSS feed and summarizes the first 5 articles in the feed.  Shows how to make API calls and parse data in a script, and call one workflow from another (workflow chaining). 
([details](./rss/README.md))

<img src="rss/rss.gif" width="800px">

## Scrape and Summarize Web content

Takes a URL and a CSS selector and scrapes web content then provides a concise summary and analysis of the text.  Shows how to use [assembllm HTML Tools wasm plug-ins](https://github.com/bradyjoslin/assembllm-htmltools) for web scraping to get clean and concise web content used in a prompt, and demonsrates a workflow built entirely using workflow chaining. ([details](./scrape_then_summarize/README.md)):

<img src="scrape_then_summarize/scrape_then_summarize.gif" width="800px">

## Weather 

Uses tools / function calling capability that takes unstructured prompt text and returns structured data, which we use to call an api to get the weather forecast for a given location. We then feed the API response into a new task, providing the original prompt and API response for context, getting a response contextualized with weather data.([details](./weather/)):

<img src="weather/weather.gif" width="800px">
