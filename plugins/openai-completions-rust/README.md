# assembllm OpenAI Rust Plug-in

[![runs_on](https://img.shields.io/badge/runs_on-Extism-4c30fc.svg?subject=runs_on&status=Extism&color=4c30fc)](https://modsurfer.dylibso.com/module?hash=114e1e892c43baefb4d50cc8b0e9f66df2b2e3177de9293ffdd83898c77e04c7)

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
