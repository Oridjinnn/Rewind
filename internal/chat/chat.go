package chat

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/habeldavidson007-glitch/rewind/internal/memory"
	"github.com/habeldavidson007-glitch/rewind/pkg/types"
)

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaStreamResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func LoadSession(
	path string,
) (types.Session, error) {

	var session types.Session

	data, err := os.ReadFile(path)

	if err != nil {
		return session, err
	}

	err = json.Unmarshal(
		data,
		&session,
	)

	if err != nil {
		return session, err
	}

	return session, nil
}

func LoadAllSessions() ([]types.Session, error) {

	var sessions []types.Session

	files, err := os.ReadDir(
		"sessions",
	)

	if err != nil {
		return sessions, err
	}

	for _, file := range files {

		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(
			file.Name(),
			".json",
		) {
			continue
		}

		path := filepath.Join(
			"sessions",
			file.Name(),
		)

		session, err := LoadSession(
			path,
		)

		if err != nil {
			fmt.Println(
				"failed loading:",
				file.Name(),
			)
			continue
		}

		sessions = append(
			sessions,
			session,
		)
	}

	return sessions, nil
}

func StartChat(model string) {
	fmt.Println("Starting chat with model:", model)
	fmt.Println("Type 'exit' to quit")
	fmt.Println("Type 'recall' to test memory recall")
	fmt.Println("")

	reader := bufio.NewReader(os.Stdin)

	var messages []types.Message
	var conversationContext string

	// Pre-load all sessions for memory recall
	sessions, err := LoadAllSessions()
	if err != nil {
		fmt.Println("Warning: Could not load sessions for memory recall:", err)
	}

	for {
		fmt.Print("You: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		input = strings.TrimSpace(input)

		if input == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		if input == "recall" {
			fmt.Println("Memory recall feature - use regular queries to search context")
			continue
		}

		if input == "" {
			continue
		}

		// Build memory context from ranked memories
		if len(sessions) > 0 {
			ranked, err := memory.RankMemories(input, sessions, 3)
			if err != nil {
				fmt.Println("Note: Memory recall failed:", err)
			} else if len(ranked) > 0 {
				fmt.Println("[Using memory context from previous sessions]")
				conversationContext = "RELEVANT CONTEXT FROM MEMORY:\n"
				for i, r := range ranked {
					conversationContext += fmt.Sprintf("%d. [Session %s] %s\n", i+1, r.SessionID, r.Content)
				}
				conversationContext += "\n---\n\n"
			}
		}

		// Build prompt with conversation history + memory context
		prompt := conversationContext
		prompt += "CONVERSATION HISTORY:\n"
		for _, msg := range messages {
			role := "USER"
			if msg.Role == "assistant" {
				role = "ASSISTANT"
			}
			prompt += fmt.Sprintf("%s: %s\n", role, msg.Content)
		}
		prompt += fmt.Sprintf("USER: %s\n", input)
		prompt += "ASSISTANT:"

		// Store user message
		userMsg := types.Message{
			Role:    "user",
			Content: input,
			Time:    time.Now(),
		}
		messages = append(messages, userMsg)

		// Stream response from Ollama
		fmt.Print("Assistant: ")
		aiResponse, err := QueryOllamaStreaming(model, prompt)
		if err != nil {
			fmt.Printf("Error: %v\n\n", err)
			continue
		}
		fmt.Println("")
		fmt.Println("")

		// Store assistant message
		assistantMsg := types.Message{
			Role:    "assistant",
			Content: aiResponse,
			Time:    time.Now(),
		}
		messages = append(messages, assistantMsg)
	}
}

// QueryOllamaStreaming sends a prompt to Ollama and streams the response in real-time
func QueryOllamaStreaming(model string, prompt string) (string, error) {
	reqBody := OllamaRequest{
		Model:  model,
		Prompt: prompt,
		Stream: true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(
		"http://localhost:11434/api/generate",
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return "", fmt.Errorf("connection error: %w (ensure Ollama is running)", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read streaming response line by line
	var fullResponse strings.Builder
	decoder := json.NewDecoder(resp.Body)

	for {
		var result OllamaStreamResponse
		err := decoder.Decode(&result)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("failed to parse response: %w", err)
		}

		// Print chunk as it arrives
		fmt.Print(result.Response)
		fullResponse.WriteString(result.Response)

		if result.Done {
			break
		}
	}

	if fullResponse.Len() == 0 {
		return "", fmt.Errorf("empty response from model")
	}

	return strings.TrimSpace(fullResponse.String()), nil
}

// QueryOllama sends a prompt to Ollama and returns the full response (non-streaming)
func QueryOllama(model string, prompt string) (string, error) {
	reqBody := OllamaRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(
		"http://localhost:11434/api/generate",
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return "", fmt.Errorf("connection error: %w (ensure Ollama is running)", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var result OllamaStreamResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Response == "" {
		return "", fmt.Errorf("empty response from model")
	}

	return strings.TrimSpace(result.Response), nil
}
