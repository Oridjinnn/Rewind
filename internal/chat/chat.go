package chat

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	"github.com/Oridjinnn/Rewind/internal/config"


	"github.com/Oridjinnn/Rewind/internal/memory"
	"github.com/Oridjinnn/Rewind/internal/storage"
	"github.com/Oridjinnn/Rewind/pkg/types"
)

var ollamaChatClient = &http.Client{
	// Timeout lebih lama untuk streaming, tapi tetap ada batasnya
	Timeout: 5 * time.Minute,
}

var saveOnce sync.Once

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
	fmt.Println("---")

	reader := bufio.NewReader(os.Stdin)
	var messages []types.Message
	var conversationContext string
	sessionID := fmt.Sprintf("chat_%d", time.Now().UnixNano())

	// Initialize Embedder for Phase 4 architecture
	embedder := memory.DefaultEmbedder()

	// Phase 1.3: Exit System Stabilization (Handle Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n[Interrupt received] Saving session and exiting...")
		if len(messages) > 0 {
			saveOnce.Do(func() {
				saveChatSession(sessionID, model, messages)
			})
		}
		os.Exit(0)
	}()

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
			ranked, err := memory.RankMemoriesV2(embedder, input, sessions, 5)
			if err != nil {
				fmt.Println("Note: Memory recall failed:", err)
			} else if len(ranked) > 0 {
				// Phase 2.4: Deduplication Logic
				uniqueContent := make(map[string]bool)
				var contextParts []string

				for _, r := range ranked {
					cleanContent := strings.TrimSpace(r.Content)
					
					// Skip if we've already added very similar content
					// or if the content is too short to be useful
					if uniqueContent[cleanContent] || len(cleanContent) < 10 {
						continue
					}
					uniqueContent[cleanContent] = true

					// Phase 2.3: Context Compression
					// Truncate long memories to keep the prompt focused
					maxSnippetLen := 400
					if len(cleanContent) > maxSnippetLen {
						cleanContent = cleanContent[:maxSnippetLen] + "... (truncated)"
					}
					
					contextParts = append(contextParts, fmt.Sprintf("[%s]: %s", r.SessionID, cleanContent))
					
					// Dynamic Budget: Max 3 high-quality memories for small LLMs
					if len(contextParts) >= 3 {
						break
					}
				}

				if len(contextParts) > 0 {
					fmt.Printf("[Cognitive Recall: %d relevant memories injected]\n", len(contextParts))
					conversationContext = "INFERRED CONTEXT FROM PREVIOUS SESSIONS:\n" + 
						strings.Join(contextParts, "\n") + 
						"\n---\nUse the context above to inform your response if relevant.\n\n"
				}
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
	saveOnce.Do(func() {
		saveChatSession(sessionID, model, messages)
	})
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

	// Phase 3.2: Reflection Engine - Inferred Cognition
	// Analyze conversation for a smarter summary
	insight, tags := ReflectOnSession(model, messages)
	summary := insight
	if len(tags) > 0 && summary != "" {
		summary = fmt.Sprintf("%s (Tags: %s)", insight, strings.Join(tags, ", "))
	}
	if summary == "" {
		summary = fmt.Sprintf("Chat with %s - %d messages", model, len(messages))
	}

	title := "Chat session"
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
		Tags:      tags,
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

	req, err := http.NewRequest("POST", config.GetOllamaHost()+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ollamaChatClient.Do(req)

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

// ReflectionResult represents structured session analysis
type ReflectionResult struct {
	Insight string   `json:"insight"`
	Tags    []string `json:"tags"`
}

// ReflectOnSession analyzes the chat to extract inferred user knowledge (Phase 3.2)
func ReflectOnSession(model string, messages []types.Message) (string, []string) {
	if len(messages) < 2 {
		return "", nil
	}

	var conv strings.Builder
	for _, m := range messages {
		conv.WriteString(fmt.Sprintf("%s: %s\n", strings.ToUpper(m.Role), m.Content))
	}

	prompt := fmt.Sprintf(`Analyze this chat session. Extract the main project goal and key technical tags (languages, tools, projects).
Return ONLY a JSON object with this format:
{"insight": "User is building...", "tags": ["golang", "sqlite", "rewind"]}

Chat:
%s

JSON:`, conv.String())

	// Phase 1.2: Timeout-safe reflection
	res, err := QueryOllama(model, prompt)
	if err != nil {
		return "", nil
	}

	// Clean JSON from potential markdown blocks
	res = strings.TrimSpace(res)
	res = strings.TrimPrefix(strings.TrimPrefix(res, "```json"), "```")
	res = strings.TrimSuffix(res, "```")
	res = strings.TrimSpace(res)

	var result ReflectionResult
	if err := json.Unmarshal([]byte(res), &result); err != nil {
		// Fallback to plain string if JSON fails
		if len(res) > 0 {
			return res, nil
		}
		return "", nil
	}

	return result.Insight, result.Tags
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

	req, err := newOllamaRequest("POST", "/api/generate", jsonData)
	if err != nil {
		return "", err
	}

	resp, err := ollamaChatClient.Do(req)

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

// newOllamaRequest builds a hardened HTTP request for the Ollama API
func newOllamaRequest(method, path string, body []byte) (*http.Request, error) {
	url := fmt.Sprintf("%s%s", config.GetOllamaHost(), path)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}