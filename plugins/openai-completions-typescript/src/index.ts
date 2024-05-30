const SUPPORTED_MODELS = [
  {"name":"gpt-4o","aliases":["4o"]},
  {"name":"gpt-4","aliases":["4"]},
  {"name":"gpt-4-1106-preview","aliases":["128k"]},
  {"name":"gpt-4-32k","aliases":["32k"]},
  {"name":"gpt-3.5-turbo","aliases":["35t"]},
  {"name":"gpt-3.5-turbo-1106","aliases":["35t-1106"]},
  {"name":"gpt-3.5-turbo-16k","aliases":["35t16k"]},
  {"name":"gpt-3.5","aliases":["35"]}
]

export function models() {
  console.log("Returning models")

  Host.outputString(JSON.stringify(SUPPORTED_MODELS))
}

function setTemperature(temperature: string): number {
  if (!temperature) {
    console.log("Temperature not set, using default value")
    return 0.7
  }
  const parsedTemperature = parseFloat(temperature)
  if (isNaN(parsedTemperature)) {
    console.log("Temperature is not a valid number, using default value")
    throw new Error("Temperature is not a valid number")
  }
  if (parsedTemperature < 0 || parsedTemperature > 1) {
    throw new Error("Temperature must be between 0 and 1")
  }
  return parsedTemperature
}

function setModel(model: string): string {
  if (!model) {
    console.log("Model not set, using default value")
    return SUPPORTED_MODELS[0].name
  }
  for (let m of SUPPORTED_MODELS) {
    if (m.name === model || m.aliases.includes(model)) {
      return m.name
    }
  }
  throw new Error("Model not found")
}

function setRole(role: string): string {
  if (!role) {
    console.log("Role not set, using default value")
    return ""
  }
  return role
}

export function completion() {
  console.log("Initiating completion")
  let apiKey = Config.get('api_key')

  if (!apiKey) {
    throw new Error("API key not set")
  } else {
    console.log("API key is set")
  }

  let temperature = setTemperature(Config.get('temperature'))
  let model = setModel(Config.get('model'))
  let role = setRole(Config.get('role'))
  let prompt = Host.inputString()

  let body = JSON.stringify({
    model,
    temperature,
    messages: [
      { role: "system", content: role },
      { role: "user", content: prompt }
    ]
  })

  let res = Http.request({
    method: "POST",
    url: "https://api.openai.com/v1/chat/completions",
    headers: {
      "Content-Type": "application/json",
      "Authorization": `Bearer ` + apiKey,
    }
    }, 
    body
  )
  let chatCompletion = JSON.parse(res.body)
  Host.outputString(chatCompletion.choices[0].message.content)
}