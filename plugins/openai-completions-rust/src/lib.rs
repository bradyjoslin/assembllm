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
struct OpenAIConfig {
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

static MODELS: [Model; 8] = [
    Model {
        name: "gpt-4o",
        aliases: ["4o"],
    },
    Model {
        name: "gpt-4",
        aliases: ["4"],
    },
    Model {
        name: "gpt-4-1106-preview",
        aliases: ["128k"],
    },
    Model {
        name: "gpt-4-32k",
        aliases: ["32k"],
    },
    Model {
        name: "gpt-3.5-turbo",
        aliases: ["35t"],
    },
    Model {
        name: "gpt-3.5-turbo-1106",
        aliases: ["35t-1106"],
    },
    Model {
        name: "gpt-3.5-turbo-16k",
        aliases: ["35t16k"],
    },
    Model {
        name: "gpt-3.5",
        aliases: ["35"],
    },
];

fn get_completion(
    api_key: String,
    model: &Model,
    input: String,
    temperature: f32,
    role: String,
) -> Result<ChatResult, anyhow::Error> {
    let req = HttpRequest::new("https://api.openai.com/v1/chat/completions")
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
    match res.status_code() {
        200 => {
            info!("Request successful");
        }
        _ => {
            let response_body = res.body();
            let body = from_utf8(&response_body)?;
            return Err(anyhow::anyhow!(
                "error calling API\nStatus Code: {}\n Response: {}",
                res.status_code(),
                body
            ));
        }
    }
    let response_body = res.body();
    let body = from_utf8(&response_body)?;
    let chat_result: ChatResult = serde_json::from_str(body)?;
    Ok(chat_result)
}

fn get_config_values(
    cfg_get: impl Fn(&str) -> Result<Option<String>, anyhow::Error>,
) -> FnResult<OpenAIConfig> {
    let api_key = cfg_get("api_key")?;
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

    Ok(OpenAIConfig {
        api_key: api_key.unwrap(),
        model: model.clone(),
        temperature,
        role,
    })
}

#[plugin_fn]
pub fn completion(input: String) -> FnResult<String> {
    let cfg = get_config_values(|key| config::get(key))?;

    let res = get_completion(cfg.api_key, &cfg.model, input, cfg.temperature, cfg.role)?;

    Ok(res.choices[0].message.content.clone())
}

#[plugin_fn]
pub fn models() -> FnResult<String> {
    let models_json = serde_json::to_string(&MODELS)?;
    info!("Returning models");
    Ok(models_json)
}
