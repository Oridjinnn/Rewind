package config

import (
	"fmt"
	"os"
)


// GetChatModel returns the Ollama model used for chat/reflection.
// Env:
// - REWIND_CHAT_MODEL
// Default: qwen2.5:1.5b
func GetChatModel() string {
	v := os.Getenv("REWIND_CHAT_MODEL")
	if v == "" {
		v = "qwen2.5:1.5b"
	}
	return v
}

// GetSummarizeModel returns the Ollama model used for summarization.
// Env:
// - REWIND_SUMMARIZE_MODEL
// Default: qwen2.5:1.5b
func GetSummarizeModel() string {
	v := os.Getenv("REWIND_SUMMARIZE_MODEL")
	if v == "" {
		v = "qwen2.5:1.5b"
	}
	return v
}

// GetEmbedModel returns the Ollama model used for embeddings.
// Env:
// - REWIND_EMBED_MODEL
// Default: nomic-embed-text
func GetEmbedModel() string {
	v := os.Getenv("REWIND_EMBED_MODEL")
	if v == "" {
		v = "nomic-embed-text"
	}
	return v
}

// GetIDEPort returns the HTTP port for the IDE event server.
// Env:
// - REWIND_IDE_PORT
// Default: 9876
func GetIDEPort() int {
	v := os.Getenv("REWIND_IDE_PORT")
	if v == "" {
		return 9876
	}
	// Parse int safely.
	// We avoid strconv.Atoi error handling by falling back to default.
	var port int
	_, err := fmt.Sscanf(v, "%d", &port)
	if err != nil || port <= 0 {
		return 9876
	}
	return port
}

