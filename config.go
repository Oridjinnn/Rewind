package config

import (
	"os"
	"strings"
)

// GetOllamaHost mengembalikan base URL Ollama dari environment variable OLLAMA_HOST atau default.
func GetOllamaHost() string {
	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		return "http://127.0.0.1:11434"
	}
	return strings.TrimSuffix(host, "/")
}