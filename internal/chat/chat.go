package chat

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/habeldavidson007-glitch/rewind/internal/memory"
	"github.com/habeldavidson007-glitch/rewind/internal/storage"
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

// StartChat starts an interactive chat session that saves to SQLite on exit.
func StartChat(model string) {
	fmt.Println("Starting chat with model:", model)
	fmt.Println("Type 'exit' to quit")
	fmt.Println("Type 'recall' to test memory recall")
	fmt.Println("")

	reader := bufio.NewReader(os.Stdin)

	var messages []types.Message
	var conversationContext string
	sessionID := fmt.Sprintf("chat_%d", time.Now().UnixNano())

	// Pre-load all sessions for memory recall using storage layer
	sessions, err := storage.LoadAllSessions()
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

	// Save chat session to SQLite storage
	saveChatSession(sessionID, model, messages)
}

// saveChatSession persists the chat conversation as a session in SQLite.
func saveChatSession(sessionID, model string, messages []types.Message) {
	if len(messages) == 0 {
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)

	// Convert messages to events
	var events []types.Event
	for _, msg := range messages {
		events = append(events, types.Event{
			Timestamp: msg.Time.UTC().Format(time.RFC3339Nano),
			Type:      msg.Role, // "user" or "assistant"
			Content:   msg.Content,
		})
	}

	// Build title from first user message
	title := "Chat session"
	summary := fmt.Sprintf("Chat with %s - %d messages", model, len(messages))
	if len(messages) > 0 {
		firstMsg := messages[0].Content
		if len(firstMsg) > 80 {
			firstMsg = firstMsg[:77] + "..."
		}
		title = firstMsg
	}

	session := types.Session{
		ID:        sessionID,
		Command:   fmt.Sprintf("rewind chat %s", model),
		Model:     model,
		Title:     title,
		Summary:   summary,
		StartedAt: now,
		EndedAt:   now,
		ExitCode:  0,
		Events:    events,
	}

	// Try to save to SQLite via getStorage pattern
	// Fall back to JSON if SQLite unavailable
	if os.Getenv("REWIND_USE_JSON") == "true" {
		saveJSONFallback(session)
		return
	}

	dbPath := os.Getenv("REWIND_DB_PATH")
	if dbPath == "" {
		dbPath = "rewind.db"
	}

	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		saveJSONFallback(session)
		return
	}
	defer store.Close()

	if err := store.SaveSession(session); err != nil {
		fmt.Printf("Note: Failed to save chat session to SQLite: %v\n", err)
		saveJSONFallback(session)
		return
	}

	fmt.Printf("Chat session saved: %s\n", sessionID)
}

// saveJSONFallback saves session as JSON when SQLite is unavailable.
func saveJSONFallback(session types.Session) {
	if err := os.MkdirAll("sessions", 0755); err != nil {
		return
	}
	path := fmt.Sprintf("sessions/%s.json", session.ID)
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(path, data, 0644)
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