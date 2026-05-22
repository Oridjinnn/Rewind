package embeddings

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Request struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type Response struct {
	Embedding []float64 `json:"embedding"`
}

func Embed(text string) ([]float64, error) {

	reqBody := Request{
		Model:  "nomic-embed-text",
		Prompt: text,
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}


	resp, err := http.Post(
		"http://127.0.0.1:11434/api/embeddings",
		"application/json",
		bytes.NewBuffer(b),
	)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out Response
	err = json.NewDecoder(resp.Body).Decode(&out)

	return out.Embedding, err
}
