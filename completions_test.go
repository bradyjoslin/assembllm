package main

import (
	"os"
	"testing"
)

func shouldSkip() bool {
    return os.Getenv("SKIP_CHAT_RESPONSE_TESTS") == "true"
}

func TestBadConfigFilePath(t *testing.T) {
	t.Parallel()

	_, err := getPluginConfig("openai", "badpath")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestGetNonExistentPluginConfig(t *testing.T) {
	t.Parallel()

	_, err := getPluginConfig("", configPath)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestGetPluginConfig(t *testing.T) {
	t.Parallel()

	want := "openai"

	pluginCfg, err := getPluginConfig(want, configPath)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	got := pluginCfg.Name

	if got != want {
		t.Fatalf("want %s, got %s", want, got)
	}
}

func TestGetModels(t *testing.T) {
	t.Parallel()

	pluginCfg, err := getPluginConfig("openai", configPath)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	models, err := pluginCfg.getModels()
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	if len(models) == 0 {
		t.Fatalf("expected models, got none")
	}
}

func TestGetResponse(t *testing.T) {
	t.Parallel()
	if shouldSkip() {
		t.Skip("Skipping this test")
	}

	pluginCfg, err := getPluginConfig("openai", configPath)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	_, err = pluginCfg.generateResponse("hello", false)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}
