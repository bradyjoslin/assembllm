# End of Life Workflow Example

**Overview**

[This workflow](./tools_eol.yaml) uses tools / function calling capability that takes unstructured prompt text and returns structured data, which we use to call an api to get product lifecycle / end of life information. We then feed the API response into a new task, providing the original prompt and API response for context, getting a response contextualized response.

**Sample Output**:

![eol gif](eol.gif)

**Usage**

```sh
assembllm -w tools_eol.yaml <prompt>
```

Example:

```
assembllm -w tools_eol.yaml 'we are using debian 9, should we think about upgrading?'
```

## Step by Step Guide

### Step 1: Fetch and Parse RSS Feed

Define an iterator script that builds a single value array with the input prompt provided.  Because an iterator's `iterValue` is retained across workflow tasks, we're using this as a means to hold the initial state of the prompt, which we'll use in our second task.  More on that below.

```yaml
iterator_script: |
  [input]
```

### Step 2:

We define a task that uses OpenAI and define a tool for that plug-in which instructs OpenAI to provide specific structured data from a given prompt, in this case find digital products from within the prompt.

```yaml
tasks:
  - name: eol
    plugin: anthropic
    tools:
      - name: eol
        description: End-of-life (EOL) and support product information
        input_schema:
          type: object
          properties:
            product:
              type: string
              description: |
                The name of a digital product.  Product may be one or more of: "akeneo-pim","alibaba-dragonwell","almalinux","alpine",...
          required:
            - product
```

At this stage a reponse to 'we are using debian 9, should we think about upgrading?' would look something like:

```json
[
  {
    "name": "eol",
    "input": {
      "product": "debian"
    }
  }
]
```

### Step 3

A post script is defined immediately after the tools call which parses the structured data response, then calls the endoflife.date api for each product found.  We limit the API data to the last 20 records to save on input tokens.  If no product is found, we return a string "no product found". 

```yaml
    post_script: |
      let jsonIn = input | fromJSON();
      map(jsonIn, {
        let product = .input.product;
        product != nil ? (Get("https://endoflife.date/api/" + product + ".json") | fromJSON() | take(20)) : "no product found"
      })
```

### Step 4

Lastly, we define another task which we'll execute a callback to OpenAI, building a prompt to let the LLM know the initial prompt from the user and the response from our tool.

```yaml
  - name: eol_response
    pre_script: |
      let primaryAsk = "The user asked: " + iterValue + ", we used a tool to find data to help answer, provide a summary response.  Here is the authoritative data: " + input;
      let noproductfound = "If no product found, just reply with only 'no product found";
      let dateContext = "First realize that today is: " + string(now()) + ", all dates should be compared against today, dates before are in the past, future hasn't happened.";
      dateContext + "\n" + primaryAsk + "\n" + noproductfound
    plugin: anthropic
```