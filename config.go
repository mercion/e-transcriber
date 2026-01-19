package main

import (
	"os"
	"strconv"
)

type Config struct {
	Addr        string
	ModelsDir   string
	ModelID     string
	ModelPath   string
	WindowSecs  int
	SampleRate  int
	PublicURL   string
	DisableDTLS bool
}

func loadConfig() Config {
	return Config{
		Addr:        getEnv("ADDR", ":8080"),
		ModelsDir:   getEnv("WHISPER_MODELS_DIR", "models"),
		ModelID:     getEnv("WHISPER_MODEL", "ggml-tiny"),
		ModelPath:   getEnv("WHISPER_MODEL_PATH", "ggml-tiny.bin"),
		WindowSecs:  getEnvInt("TRANSCRIBE_WINDOW_SECS", 5),
		SampleRate:  whisperSampleRate,
		PublicURL:   getEnv("PUBLIC_URL", ""),
		DisableDTLS: getEnvBool("DISABLE_DTLS", false),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.Atoi(value)
		if err == nil {
			return parsed
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err == nil {
			return parsed
		}
	}
	return fallback
}
