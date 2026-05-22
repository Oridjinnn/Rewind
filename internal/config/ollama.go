package config

import "os"

// GetOllamaHost returns base URL for the Ollama server.
// Env:
// - REWIND_OLLAMA_HOST (full base URL, e.g. http://127.0.0.1:11434)
// Default: http://localhost:11434
func GetOllamaHost() string {
	host := os.Getenv("REWIND_OLLAMA_HOST")
	if host == "" {
		host = "http://localhost:11434"
	}
	return host
}
