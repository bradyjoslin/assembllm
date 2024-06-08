package main

import "strings"

func updatePlugins(configData []byte) []byte {
	currentConfig := string(configData)
	configUpdates := currentConfig
	// Mapping of old plug-in hashes to latest
	updates := []struct {
		Old []struct {
			Source, Hash string
		}
		New struct {
			Source, Hash string
		}
	}{
		{
			Old: []struct{ Source, Hash string }{
				{
					Source: "https://cdn.modsurfer.dylibso.com/api/v1/module/114e1e892c43baefb4d50cc8b0e9f66df2b2e3177de9293ffdd83898c77e04c7.wasm",
					Hash:   "114e1e892c43baefb4d50cc8b0e9f66df2b2e3177de9293ffdd83898c77e04c7",
				},
				{
					Source: "https://cdn.modsurfer.dylibso.com/api/v1/module/e5768c2835a01ee1a5f10702020a82e0ba2166ba114733e2215b2c2ef423985f.wasm",
					Hash:   "e5768c2835a01ee1a5f10702020a82e0ba2166ba114733e2215b2c2ef423985f",
				},
			},
			New: struct{ Source, Hash string }{
				Source: "https://github.com/bradyjoslin/assembllm-openai/releases/latest/download/assembllm_openai.wasm",
				Hash:   "",
			},
		},
		{
			Old: []struct{ Source, Hash string }{
				{
					Source: "https://cdn.modsurfer.dylibso.com/api/v1/module/dd58ff133011b296ff5ba00cc3b0b4df34c1a176e5aebff9643d1ac83b88c72b.wasm",
					Hash:   "dd58ff133011b296ff5ba00cc3b0b4df34c1a176e5aebff9643d1ac83b88c72b",
				},
			},
			New: struct{ Source, Hash string }{
				Source: "https://github.com/bradyjoslin/assembllm-cloudflare/releases/latest/download/assembllm_cloudflare.wasm",
				Hash:   "",
			},
		},
		{
			Old: []struct{ Source, Hash string }{
				{
					Source: "https://cdn.modsurfer.dylibso.com/api/v1/module/9c1a87483040d5033866fc5b8581cc8aa7bc18abd9a601a14a4dec998a5a75f9.wasm",
					Hash:   "9c1a87483040d5033866fc5b8581cc8aa7bc18abd9a601a14a4dec998a5a75f9",
				},
			},
			New: struct{ Source, Hash string }{
				Source: "https://github.com/bradyjoslin/assembllm-perplexity/releases/latest/download/assembllm_perplexity.wasm",
				Hash:   "",
			},
		},
		{
			Old: []struct{ Source, Hash string }{
				{
					Source: "https://cdn.modsurfer.dylibso.com/api/v1/module/93f3517589bd44dfde3a0406ab2d574f239aca10378996bb6c63e8d73a510e2b.wasm",
					Hash:   "93f3517589bd44dfde3a0406ab2d574f239aca10378996bb6c63e8d73a510e2b",
				},
			},
			New: struct{ Source, Hash string }{
				Source: "https://github.com/bradyjoslin/assembllm-openai-go/releases/latest/download/assembllm-openai-go.wasm",
				Hash:   "",
			},
		},
		{
			Old: []struct{ Source, Hash string }{
				{
					Source: "https://cdn.modsurfer.dylibso.com/api/v1/module/6d2e458bf3eea4925503bc7803c0d01366430a8e2779bd088b8f9887745b4e00.wasm",
					Hash:   "6d2e458bf3eea4925503bc7803c0d01366430a8e2779bd088b8f9887745b4e00",
				},
			},
			New: struct{ Source, Hash string }{
				Source: "https://github.com/bradyjoslin/assembllm-openai-csharp/releases/latest/download/assembllm-openai-csharp.wasm",
				Hash:   "",
			},
		},
		{
			Old: []struct{ Source, Hash string }{
				{
					Source: "https://cdn.modsurfer.dylibso.com/api/v1/module/a9110e703ff5c68cbf028c725851fd287ac1ef0b909b1d97c600f881e272fa8c.wasm",
					Hash:   "a9110e703ff5c68cbf028c725851fd287ac1ef0b909b1d97c600f881e272fa8c",
				},
			},
			New: struct{ Source, Hash string }{
				Source: "https://github.com/bradyjoslin/assembllm-openai-ts/releases/latest/download/assembllm-openai-ts.wasm",
				Hash:   "",
			},
		},
	}

	// Check if updates are needed
	for _, update := range updates {
		for _, old := range update.Old {
			// Replace the old source and hash with the new ones
			configUpdates = strings.Replace(configUpdates, old.Source, update.New.Source, -1)
			configUpdates = strings.Replace(configUpdates, old.Hash, update.New.Hash, -1)
		}
	}

	return []byte(configUpdates)
}
