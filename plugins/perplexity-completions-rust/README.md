# assembllm Perplexity Rust Plug-in

[![runs_on](https://img.shields.io/badge/runs_on-Extism-4c30fc.svg?subject=runs_on&status=Extism&color=4c30fc)](https://modsurfer.dylibso.com/module?hash=9c1a87483040d5033866fc5b8581cc8aa7bc18abd9a601a14a4dec998a5a75f9)

Uses [Extism PDK](https://github.com/extism/rust-pdk).  Doesn't require WASI.

## Building

```bash
make build
```

Built wasm file will be in `target/wasm32-unknown-unknown/release/`.

## Testing

```bash
make test
```
