CF_ACCOUNT_ID := $(shell echo $$CF_ACCOUNT_ID)
WORKERS_AI_TOKEN := $(shell echo $$WORKERS_AI_TOKEN)

build:
	cargo build --release --target wasm32-unknown-unknown

test:
	extism call ./target/wasm32-unknown-unknown/release/cloudflare_completions.wasm models --log-level=info
	@extism call ./target/wasm32-unknown-unknown/release/cloudflare_completions.wasm completion \
		--set-config='{"api_key": "$(WORKERS_AI_TOKEN)", "account_id": "$(CF_ACCOUNT_ID)"}' \
		--input="Explain extism in the context of wasm succinctly." \
		--allow-host=api.cloudflare.com --log-level=info