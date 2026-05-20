package summarize

import (
	"bytes"
	"encoding/json"
	"net/http"

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

func GenerateSummary(
	conversation string,
) string {

	prompt := `
You are a memory summarizer.

Summarize this conversation in 1 concise sentence.

Conversation:
` + conversation

	reqBody := OllamaRequest{
		Model:  "qwen2.5:1.5b",
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
		return "Summary generation failed."
	}

	defer resp.Body.Close()

	var result OllamaResponse

	json.NewDecoder(resp.Body).Decode(&result)

	return result.Response
}
