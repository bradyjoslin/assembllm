# RSS Workflow Example

**Overview**

[This workflow](./tools_weather.yaml) Uses tools / function calling capability that takes unstructured prompt text and returns structured data, which we use to call an api to get the weather forecast for a given location. We then feed the API response into a new task, providing the original prompt and API response for context, getting a response contextualized with weather data.

**Sample Output**:

![weather](weather.gif)

**Usage**

```sh
assembllm -w tools_weather.yaml <prompt>
```

Example:

```
assembllm -w ~/Projects/assembllm/workflows/weather/tools_weather.yaml "heading to the beach tomorrow in grand cayman, should i expect a full day of sun?"
```

## Step by Step Example

### Step 1: 

Define an iterator script that builds a single value array with the input prompt provided.  Because an iterator's `iterValue` is retained across workflow tasks, we're using this as a means to hold the initial state of the prompt, which we'll use in our second task.  More on that below.

```yaml
iterator_script: |
  [input]
```

### Step 2:

We define a task that uses OpenAI and define a tool for that plug-in which instructs OpenAI to provide specific structured data from a given prompt, in this case location and temperature units.

```yaml
tasks:
  - name: weather
    plugin: openai
    tools:
      - name: weather
        description: Get the current weather
        input_schema:
          type: object
          properties:
            location:
              type: string
              description: The city and state, e.g. San Francisco CA
            units:
              type: string
              description: The temperature unit to use. Infer this from the users location. e.g. F or C.
          required:
            - location
            - units
```

At this stage a reponse to "driving from houston to hot springs today, interested in expected driving conditions" would look something like:

```json
[
  {
    "input": {
      "location": "Houston, TX",
      "units": "F"
    },
    "name": "weather"
  },
  {
    "input": {
      "location": "Hot Springs, AR",
      "units": "F"
    },
    "name": "weather"
  }
]
```

### Step 3

A post script is defined immediately after the tools call which parses the structured data response.  We use an [Expr map function](https://expr-lang.org/docs/language-definition#map) because we receive a JSON array response from OpenAPI, which allows handling responses that contain multiple locations.

```yaml
    post_script: |
      let jsonIn = input | fromJSON();
      map(jsonIn, {
        let location = .input.location;
        let units = .input.units;
        let formattedLocation = replace(location, " ", "+") | replace(",", "");
        let unitOption = units == "F" ? "u" : "";
        Get("https://wttr.in/" + formattedLocation + "?dA" + unitOption)
      })
```

This gets the weather report for that location from [wttr.in](https://wttr.in), requesting Fahrenheit, when appropriate.

### Step 4

Lastly, we define another task which we'll execute a callback to OpenAI, building a prompt to let it know the initial prompt from the user and the response from our tool, the weather service.

```yaml
- name: weather_response
    plugin: openai
    pre_script: |
      "The user asked: " + iterValue + ", we used a tool to find data to help answer, please summarize for them: " + input
```
