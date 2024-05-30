OPENAI_API_KEY := $(shell echo $$OPENAI_API_KEY)

build:
	tinygo build -target wasi -o plugin.wasm main.go

test:
	extism call plugin.wasm models --input "world" --wasi --log-level info
	@extism call plugin.wasm completion --input "world" --wasi --allow-host "api.openai.com" --log-level info --set-config='{"api_key": "$(OPENAI_API_KEY)"}'