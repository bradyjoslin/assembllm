use extism_pdk::*;
use serde::{Deserialize, Serialize};
use serde_json::json;
use std::str::from_utf8;

#[derive(Debug, Deserialize)]
struct ChatMessage {
    content: String,
}

#[derive(Debug, Deserialize)]
struct ChatChoice {
    message: ChatMessage,
}

#[derive(Debug, Deserialize)]
struct ChatResult {
    choices: Vec<ChatChoice>,
}

#[derive(Debug)]
struct CloudflareAIConfig {
    account_id: String,
    api_key: String,
    model: Model,
    temperature: f32,
    role: String,
}

#[derive(Clone, Debug, Serialize)]
struct Model {
    name: &'static str,
    aliases: [&'static str; 1],
}

static MODELS: [Model; 10] = [
    Model {
        name: "@cf/meta/llama-3-8b-instruct",
        aliases: ["llama-3-8b"],
    },
    Model {
        name: "@cf/meta/llama-2-7b-chat-fp16",
        aliases: ["llama-2-7b"],
    },
    Model {
        name: "@cf/meta/llama-2-7b-chat-int8",
        aliases: ["llama-2-7b-int8"],
    },
    Model {
        name: "@cf/mistral/mistral-7b-instruct-v0.1",
        aliases: ["mistral-7b"],
    },
    Model {
        name: "@hf/thebloke/deepseek-coder-6.7b-base-awq",
        aliases: ["deepseek-coder-6.7b"],
    },
    Model {
        name: "@hf/thebloke/deepseek-coder-6.7b-instruct-awq",
        aliases: ["deepseek-coder-6.7b-instruct"],
    },
    Model {
        name: "@cf/deepseek-ai/deepseek-math-7b-base",
        aliases: ["deepseek-math-7b"],
    },
    Model {
        name: "@cf/deepseek-ai/deepseek-math-7b-instruct",
        aliases: ["deepseek-math-7b-instruct"],
    },
    Model {
        name: "@cf/tiiuae/falcon-7b-instruct",
        aliases: ["falcon-7b"],
    },
    Model {
        name: "@cf/google/gemma-2b-it-lora",
        aliases: ["gemma-2b"],
    },
];

fn get_completion(
    api_key: String,
    model: &Model,
    input: String,
    temperature: f32,
    role: String,
    account_id: String,
) -> Result<ChatResult, anyhow::Error> {
    let req = HttpRequest::new(format!("https://api.cloudflare.com/client/v4/accounts/{}/ai/v1/chat/completions", account_id))
        .with_header("Authorization", format!("Bearer {}", api_key))
        .with_header("Content-Type", "application/json")
        .with_method("POST");

    // We could make our own structs for the body
    // this is a quick way to make some unstructured JSON
    let req_body = json!({
      "model": model.name,
      "temperature": temperature,
      "messages": [
        {
            "role": "system",
            "content": role,
          },
        {
          "role": "user",
          "content": input,
        }
      ],
    });

    let res = http::request::<String>(&req, Some(req_body.to_string()))?;
    let body = res.body();
    let body = from_utf8(&body)?;

    let body: ChatResult = serde_json::from_str(body)?;
    Ok(body)
}

fn get_config_values(
    cfg_get: impl Fn(&str) -> Result<Option<String>, anyhow::Error>,
) -> FnResult<CloudflareAIConfig> {
    let api_key = cfg_get("api_key")?;
    let account_id = cfg_get("account_id")?;
    let model_input = cfg_get("model")?;
    let temperature_input = cfg_get("temperature")?;
    let role_input = cfg_get("role")?;

    match api_key {
        Some(_) => {
            info!("API key found");
        }
        None => {
            error!("API key not found");
            return Err(WithReturnCode::new(anyhow::anyhow!("API key not found"), 1));
        }
    }

    let model = match model_input {
        Some(model) => {
            let found_model = MODELS.iter().find(|m| {
                m.name.to_lowercase() == model.to_lowercase()
                    || m.aliases
                        .iter()
                        .any(|&alias| alias.to_lowercase() == model.to_lowercase())
            });
            match found_model {
                Some(m) => {
                    info!("Model found: {}", m.name);
                    m
                }
                None => {
                    error!("Model not found");
                    return Err(WithReturnCode::new(anyhow::anyhow!("Model not found"), 1));
                }
            }
        }
        _ => {
            info!("Model not specified, using default");
            MODELS.first().unwrap()
        }
    };

    let temperature = match temperature_input {
        Some(temperature) => {
            let t = temperature.parse::<f32>();
            match t {
                Ok(t) => {
                    if t < 0.0 || t > 1.0 {
                        error!("Temperature must be between 0.0 and 1.0");
                        return Err(WithReturnCode::new(
                            anyhow::anyhow!("Temperature must be between 0.0 and 1.0"),
                            1,
                        ));
                    }
                    info!("Temperature: {}", t);
                    t
                }
                Err(_) => {
                    error!("Temperature must be a float");
                    return Err(WithReturnCode::new(
                        anyhow::anyhow!("Temperature must be a float"),
                        1,
                    ));
                }
            }
        }
        None => {
            info!("Temperature not specified, using default");
            0.7
        }
    };

    let role = role_input.unwrap_or("".to_string());
    if role != "" {
        info!("Role: {}", role);
    } else {
        info!("Role not specified");
    }

    Ok(CloudflareAIConfig {
        account_id: account_id.unwrap(),
        api_key: api_key.unwrap(),
        model: model.clone(),
        temperature,
        role,
    })
}

#[plugin_fn]
pub fn completion(input: String) -> FnResult<String> {
    let cfg = get_config_values(|key| config::get(key))?;

    let res = get_completion(cfg.api_key, &cfg.model, input, cfg.temperature, cfg.role, cfg.account_id)?;

    Ok(res.choices[0].message.content.clone())
}

#[plugin_fn]
pub fn models() -> FnResult<String> {
    let models_json = serde_json::to_string(&MODELS)?;
    info!("Returning models");
    Ok(models_json)
}
