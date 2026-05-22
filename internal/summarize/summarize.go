package summarize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Oridjinnn/Rewind/internal/config"


	"github.com/Oridjinnn/Rewind/pkg/types"
)

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

// SummaryData stores structured summary information
type SummaryData struct {
	Title      string   `json:"title"`
	OneLiners  string   `json:"one_liners"`
	KeyPoints  []string `json:"key_points"`
	Entities   []string `json:"entities"`
	Sentiment  string   `json:"sentiment"`
	Duration   string   `json:"duration"`
	EventCount int      `json:"event_count"`
}

func BuildConversation(
	events []types.Event,
) string {

	var out string

	for _, e := range events {

		if e.Type == "user_message" {
			out += "USER: " + e.Content + "\n"
		}

		if e.Type == "assistant_message" {
			out += "ASSISTANT: " + e.Content + "\n"
		}
	}

	return out
}

// GenerateSummary creates a structured summary of the session
// Merges 4 separate Ollama calls into 1 structured prompt (4x faster).
func GenerateSummary(conversation string) string {
	prompt := fmt.Sprintf(`Analyze this conversation and return a JSON object with these fields:
{
  "one_liner": "1 sentence summary (max 15 words)",
  "key_points": ["point1", "point2", "point3"],
  "entities": ["entity1", "entity2"],
  "sentiment": "positive|negative|neutral|confused|excited|frustrated"
}

Return ONLY the JSON, no other text.

Conversation:
%s`, conversation)

	model := os.Getenv("REWIND_SUMMARIZE_MODEL")
	if model == "" {
		model = "qwen2.5:1.5b"
	}

	response := QueryOllamaSimple(model, prompt)

	// Parse structured JSON response
	var parsed struct {
		OneLiner  string   `json:"one_liner"`
		KeyPoints []string `json:"key_points"`
		Entities  []string `json:"entities"`
		Sentiment string   `json:"sentiment"`
	}

	if err := json.Unmarshal([]byte(response), &parsed); err != nil {
		// Fallback: treat whole response as summary if JSON parse fails
		parsed.OneLiner = response
	}

	summary := SummaryData{
		Title:      parsed.OneLiner,
		OneLiners:  parsed.OneLiner,
		KeyPoints:  parsed.KeyPoints,
		Entities:   parsed.Entities,
		Sentiment:  parsed.Sentiment,
		EventCount: strings.Count(conversation, "USER:") + strings.Count(conversation, "ASSISTANT:"),
	}

	result := fmt.Sprintf(`Summary: %s
Key Points: %s
Sentiment: %s
Entities: %s`,
		summary.Title,
		strings.Join(summary.KeyPoints, ", "),
		summary.Sentiment,
		strings.Join(summary.Entities, ", "),
	)

	return result
}

// QueryOllamaSimple is a helper for simple Ollama queries
func QueryOllamaSimple(model string, prompt string) string {

	reqBody := OllamaRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "Error marshalling request"
	}

	resp, err := http.Post(
		config.GetOllamaHost()+"/api/generate",
		"application/json",
		bytes.NewBuffer(jsonData),
	)


	if err != nil {
		return "Error connecting to Ollama"
	}

	defer resp.Body.Close()

	var result OllamaResponse

	json.NewDecoder(resp.Body).Decode(&result)

	return result.Response
}
