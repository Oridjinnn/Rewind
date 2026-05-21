package summarize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/habeldavidson007-glitch/rewind/pkg/types"
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
func GenerateSummary(conversation string) string {
	
	// Extract key points
	keyPoints := ExtractKeyPoints(conversation)
	
	// Extract entities
	entities := ExtractEntities(conversation)
	
	// Generate one-liner
	oneLiner := GenerateOneLiner(conversation)
	
	// Analyze sentiment
	sentiment := AnalyzeSentiment(conversation)

	summary := SummaryData{
		Title:      oneLiner,
		OneLiners:  oneLiner,
		KeyPoints:  keyPoints,
		Entities:   entities,
		Sentiment:  sentiment,
		EventCount: strings.Count(conversation, "USER:") + strings.Count(conversation, "ASSISTANT:"),
	}

	// Return as formatted text for display
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

// ExtractKeyPoints uses Ollama to extract main topics
func ExtractKeyPoints(conversation string) []string {
	
	prompt := `Extract 3-5 key points from this conversation. Return only the points, one per line.

Conversation:
` + conversation

	response := QueryOllamaSimple("qwen2.5:1.5b", prompt)
	
	lines := strings.Split(response, "\n")
	var points []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "-") {
			points = append(points, line)
		} else if strings.HasPrefix(line, "-") {
			points = append(points, strings.TrimPrefix(strings.TrimSpace(line), "- "))
		}
	}

	if len(points) > 5 {
		points = points[:5]
	}

	return points
}

// ExtractEntities finds named entities in conversation
func ExtractEntities(conversation string) []string {
	
	prompt := `List important named entities (people, places, projects, tools) from this conversation. Return only the names, comma-separated.

Conversation:
` + conversation

	response := QueryOllamaSimple("qwen2.5:1.5b", prompt)
	
	entities := strings.Split(response, ",")
	var result []string
	for _, e := range entities {
		e = strings.TrimSpace(e)
		if e != "" {
			result = append(result, e)
		}
	}

	if len(result) > 5 {
		result = result[:5]
	}

	return result
}

// AnalyzeSentiment determines overall tone of conversation
func AnalyzeSentiment(conversation string) string {
	
	prompt := `In one word, describe the sentiment of this conversation (positive, negative, neutral, confused, excited, frustrated, etc):

Conversation:
` + conversation

	response := QueryOllamaSimple("qwen2.5:1.5b", prompt)
	
	return strings.TrimSpace(response)
}

// GenerateOneLiner creates a concise summary sentence
func GenerateOneLiner(conversation string) string {
	
	prompt := `Summarize this conversation in exactly 1 concise sentence (max 15 words):

Conversation:
` + conversation

	response := QueryOllamaSimple("qwen2.5:1.5b", prompt)
	
	return strings.TrimSpace(response)
}

// QueryOllamaSimple is a helper for simple Ollama queries
func QueryOllamaSimple(model string, prompt string) string {

	reqBody := OllamaRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post(
		"http://localhost:11434/api/generate",
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
