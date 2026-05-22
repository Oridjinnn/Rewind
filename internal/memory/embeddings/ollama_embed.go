package embeddings

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/Oridjinnn/Rewind/internal/config"
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
		Model:  config.GetEmbedModel(),
		Prompt: text,
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}


	resp, err := http.Post(
		config.GetOllamaHost()+"/api/embeddings",
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
