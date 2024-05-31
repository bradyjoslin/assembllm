# assembllm Cloudflare Rust Plug-in

[![runs_on](https://img.shields.io/badge/runs_on-Extism-4c30fc.svg?subject=runs_on&status=Extism&color=4c30fc)](https://modsurfer.dylibso.com/module?hash=dd58ff133011b296ff5ba00cc3b0b4df34c1a176e5aebff9643d1ac83b88c72b)

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
