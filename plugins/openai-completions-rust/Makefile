OPENAI_API_KEY := $(shell echo $$OPENAI_API_KEY)

build:
	cargo build --release --target wasm32-unknown-unknown

test:
	extism call ./target/wasm32-unknown-unknown/release/openai_completions.wasm models --log-level=info
	@extism call ./target/wasm32-unknown-unknown/release/openai_completions.wasm completion \
		--set-config='{"api_key": "$(OPENAI_API_KEY)"}' \
		--input="Explain extism in the context of wasm succinctly." \
		--allow-host=api.openai.com --log-level=info