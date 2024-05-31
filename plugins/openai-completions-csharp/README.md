# assembllm OpenAI C# Plug-in

[![runs_on](https://img.shields.io/badge/runs_on-Extism-4c30fc.svg?subject=runs_on&status=Extism&color=4c30fc)](https://modsurfer.dylibso.com/module?hash=6d2e458bf3eea4925503bc7803c0d01366430a8e2779bd088b8f9887745b4e00) [![standard](https://img.shields.io/badge/standard-WASI%20(preview1)-654ff0.svg?subject=standard&status=WASI%20(preview1)&color=654ff0)](https://modsurfer.dylibso.com/module?hash=6d2e458bf3eea4925503bc7803c0d01366430a8e2779bd088b8f9887745b4e00)

Uses [Extism PDK](https://github.com/extism/dotnet-pdk).  Requires WASI.

## Building

Debug build:

```bash
make build
```

Built wasm file will be in `bin/Debug/net8.0/wasi-wasm/AppBundle`.

Release Build:

```bash
make release
```

Built wasm file will be in `bin/Release/net8.0/wasi-wasm/AppBundle/`.

## Testing

Test debug:

```bash
make test_build
```

Test release:

```bash
make test_release
```
