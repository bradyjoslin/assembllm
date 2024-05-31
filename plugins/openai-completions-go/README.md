# assembllm OpenAI Go Plug-in

[![runs_on](https://img.shields.io/badge/runs_on-Extism-4c30fc.svg?subject=runs_on&status=Extism&color=4c30fc)](https://modsurfer.dylibso.com/module?hash=93f3517589bd44dfde3a0406ab2d574f239aca10378996bb6c63e8d73a510e2b) [![standard](https://img.shields.io/badge/standard-WASI%20(preview1)-654ff0.svg?subject=standard&status=WASI%20(preview1)&color=654ff0)](https://modsurfer.dylibso.com/module?hash=93f3517589bd44dfde3a0406ab2d574f239aca10378996bb6c63e8d73a510e2b)

Requires [tinygo](https://tinygo.org/).

Uses [Extism PDK](https://github.com/extism/go-pdk). Requires WASI.

## Building

```bash
make build
```

Built wasm file will be in `target/wasm32-unknown-unknown/release/`.

## Testing

```bash
make test
```
