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

	"github.com/Oridjinnn/Rewind/pkg/types"
)

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
func GetCachePath(sessionID string) string {
	return filepath.Join(".rewind", "embeddings", fmt.Sprintf("%s.json", sessionID))
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

func EmbedText(text string) ([]float64, error) {

	if text == "" {
		return nil, errors.New("empty text")
	}

	payload := map[string]string{
		"model":  "nomic-embed-text",
		"prompt": text,
	}

	body, _ := json.Marshal(payload)

	resp, err := http.Post(
		"http://127.0.0.1:11434/api/embeddings",
		"application/json",
		bytes.NewBuffer(body),
	)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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
	// Ensure vectors are the same length and not empty
	if len(a) != len(b) || len(a) == 0 {
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
	Score     float64
}

func RankMemories(query string, sessions []types.Session, topK int) ([]RankedMemory, error) {

	queryVec, err := EmbedText(query)
	if err != nil {
		return nil, err
	}

	var ranked []RankedMemory

	for _, s := range sessions {
		// Load cache for this session
		cache, err := LoadEmbeddingCache(s.ID)
		if err != nil {
			// Log but continue - cache errors shouldn't break ranking
			fmt.Printf("Warning: failed to load cache for session %s: %v\n", s.ID, err)
		}

		for eventIdx, e := range s.Events {

			if e.Type != "user_message" && e.Type != "assistant_message" {
				continue
			}

			// Get or embed - uses cache if available
			vec, err := GetOrEmbedEvent(s.ID, eventIdx, e, cache)
			if err != nil {
				continue
			}

			score := CosineSimilarity(queryVec, vec)

			if score == 0 {
				continue
			}

			ranked = append(ranked, RankedMemory{
				SessionID: s.ID,
				Content:   e.Content,
				Score:     score,
			})
		}

		// Save updated cache (with any newly embedded events)
		if cache != nil && len(cache.Events) > 0 {
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
