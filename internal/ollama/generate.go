package ollama

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type GenerateResponse struct {
	Response string `json:"response"`
}

func Generate(
	model string,
	prompt string,
) (string, error) {

	reqBody := GenerateRequest{
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
		return "", err
	}

	defer resp.Body.Close()

	var result GenerateResponse

	err = json.NewDecoder(resp.Body).Decode(
		&result,
	)

	if err != nil {
		return "", err
	}

	return result.Response, nil
}
