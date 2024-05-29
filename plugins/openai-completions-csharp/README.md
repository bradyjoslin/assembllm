# assembllm OpenAI C# Plug-in

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
