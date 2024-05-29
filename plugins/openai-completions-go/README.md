# assembllm OpenAI Go Plug-in

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
