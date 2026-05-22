package memory

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/habeldavidson007-glitch/rewind/pkg/types"
)

// Global http client with timeout to prevent hanging
var ollamaClient = &http.Client{
	Timeout: 10 * time.Second,
}

// Embedder defines the interface for generating vector embeddings.
type Embedder interface {
	EmbedText(text string) ([]float64, error)
}

// OllamaEmbedder implements the Embedder interface using Ollama API.
type OllamaEmbedder struct {
	BaseURL string
	Model   string
	Client  *http.Client
}

type EmbedResponse struct {
	Embedding []float64 `json:"embedding"`
}

// Cache structures for embeddings
type EventEmbedding struct {
	EventIndex int       `json:"event_index"`
	Type       string    `json:"type"`
	Content    string    `json:"content"`
	Embedding  []float64 `json:"embedding"`
	Timestamp  string    `json:"timestamp"`
}

type EmbeddingCache struct {
	SessionID string             `json:"session_id"`
	Events    []EventEmbedding   `json:"events"`
}

// GetCachePath returns the path to the embedding cache file for a session.
// Embedding cache is stored in .rewind/embeddings/ (separated from user session data).
// NOTE: To be deprecated in v0.3.1 in favor of SQLite embedding table.
func GetCachePath(sessionID string) string {
	dir := filepath.Join(".rewind", "embeddings")
	// Ensure directory exists
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, fmt.Sprintf("%s.json", sessionID))
}

// LoadEmbeddingCache loads cached embeddings from disk
func LoadEmbeddingCache(sessionID string) (*EmbeddingCache, error) {
	cachePath := GetCachePath(sessionID)

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Cache doesn't exist yet
		}
		return nil, err
	}

	var cache EmbeddingCache
	err = json.Unmarshal(data, &cache)
	if err != nil {
		return nil, err
	}

	return &cache, nil
}

// SaveEmbeddingCache saves embeddings to disk
func SaveEmbeddingCache(cache *EmbeddingCache) error {
	cachePath := GetCachePath(cache.SessionID)

	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	err = os.WriteFile(cachePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// GetOrEmbedEvent gets embedding from cache or generates it
func GetOrEmbedEvent(sessionID string, eventIndex int, event types.Event, cache *EmbeddingCache) ([]float64, error) {
	
	// Check cache first
	if cache != nil {
		for _, e := range cache.Events {
			if e.EventIndex == eventIndex && e.Content == event.Content {
				return e.Embedding, nil
			}
		}
	}

	// Not in cache, embed it
	vec, err := EmbedText(event.Content)
	if err != nil {
		return nil, err
	}

	// Add to cache
	if cache != nil {
		cache.Events = append(cache.Events, EventEmbedding{
			EventIndex: eventIndex,
			Type:       event.Type,
			Content:    event.Content,
			Embedding:  vec,
			Timestamp:  event.Timestamp,
		})
	}

	return vec, nil
}

// DefaultEmbedder returns a pre-configured Ollama embedder.
func DefaultEmbedder() Embedder {
	return &OllamaEmbedder{
		BaseURL: "http://127.0.0.1:11434",
		Model:   "nomic-embed-text",
		Client:  ollamaClient,
	}
}

func (o *OllamaEmbedder) EmbedText(text string) ([]float64, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, errors.New("empty text")
	}

	payload := map[string]string{
		"model":  o.Model,
		"prompt": trimmed,
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", o.BaseURL+"/api/embeddings", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama embedding failed with status: %d", resp.StatusCode)
	}

	var result EmbedResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	if len(result.Embedding) == 0 {
		return nil, errors.New("empty embedding returned")
	}

	return result.Embedding, nil
}

func CosineSimilarity(a, b []float64) float64 {
	if a == nil || b == nil || len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dot, normA, normB float64
	for i := 0; i < len(a); i++ {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

type RankedMemory struct {
	SessionID string
	Content   string
	Type      string
	Score     float64
}

// RankMemoriesV2 implements Hybrid Ranking: Semantic Similarity + Recency Decay.
func RankMemoriesV2(embedder Embedder, query string, sessions []types.Session, topK int) ([]RankedMemory, error) {
	queryVec, err := embedder.EmbedText(query)
	if err != nil {
		return nil, err
	}

	var ranked []RankedMemory

	for _, s := range sessions {
		// Phase 1.1: Load cache for this session
		cache, err := LoadEmbeddingCache(s.ID)
		if err != nil {
			// Log but continue - cache errors shouldn't break ranking
			fmt.Printf("Warning: failed to load cache for session %s: %v\n", s.ID, err)
		}

		// Initialize empty cache object if none exists to enable population
		if cache == nil {
			cache = &EmbeddingCache{SessionID: s.ID, Events: []EventEmbedding{}}
		}

		dirtyCache := false
		for eventIdx, e := range s.Events {

			// Phase 2.2: Prioritaskan pesan user dan assistant, abaikan event sistem lainnya
			if e.Type != "user" && e.Type != "assistant" && e.Type != "user_message" && e.Type != "assistant_message" {
				continue
			}

			// Skip pesan yang terlalu pendek yang biasanya tidak memiliki konteks semantik yang kuat
			if len(strings.TrimSpace(e.Content)) < 10 {
				continue
			}

			// Get or embed - uses cache if available
			vec, err := GetOrEmbedEvent(s.ID, eventIdx, e, cache)
			if err != nil {
				continue
			}

			// Phase 2.1: Semantic Similarity
			semanticScore := CosineSimilarity(queryVec, vec)

			// Phase 2.2: Relevance Threshold
			// Only keep memories that are actually relevant (> 0.45)
			if semanticScore < 0.45 {
				continue
			}

			// Phase 2.1: Recency Score (30% weight)
			recencyScore := calculateRecencyScore(e.Timestamp)

			// Phase 2.2: Importance Heuristic
			// Pesan dari user biasanya berisi instruksi atau konteks yang lebih krusial untuk recall
			importance := 1.0
			if e.Type == "user" || e.Type == "user_message" {
				importance = 1.2
			}

			finalScore := ((semanticScore * 0.7) + (recencyScore * 0.3)) * importance

			ranked = append(ranked, RankedMemory{
				SessionID: s.ID,
				Content:   e.Content,
				Type:      e.Type,
				Score:     finalScore,
			})
		}

		// Save updated cache (with any newly embedded events)
		if dirtyCache {
			err = SaveEmbeddingCache(cache)
			if err != nil {
				fmt.Printf("Warning: failed to save cache for session %s: %v\n", s.ID, err)
			}
		}
	}

	// Sort by score descending (O(n log n))
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Score > ranked[j].Score
	})

	if len(ranked) > topK {
		ranked = ranked[:topK]
	}

	return ranked, nil
}

// findInCache helper to look up existing embeddings
func findInCache(cache *EmbeddingCache, idx int, content string) []float64 {
	if cache == nil {
		return nil
	}
	for _, e := range cache.Events {
		if e.EventIndex == idx && e.Content == content {
			return e.Embedding
		}
	}
	return nil
}

func calculateRecencyScore(timestamp string) float64 {
	// RFC3339 handles sub-second precision (Nano) gracefully
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return 0.5
	}
	daysOld := time.Since(t).Hours() / 24.0
	// Decay formula: 1 / (1 + days/7). Memori 7 hari yang lalu punya skor 0.5
	return 1.0 / (1.0 + (daysOld / 7.0))
}
