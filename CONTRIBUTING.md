# Contributing to Rewind

First off: thanks for taking the time to contribute! ❤️

## How to contribute

### 1) Get the repo running
- Ensure you have **Go 1.22+** installed.
- Ensure **Ollama** is running locally (for AI features).

### 2) Choose an issue or propose one
- Look for issues (bug reports, feature requests).
- If none exist, open a new issue describing what you want to build and why.

### 3) Make changes
- Keep changes small and focused.
- Follow existing code style.
- Prefer deterministic behavior for tests.

### 4) Test
Run:
- `go test ./...`
- `go vet ./...`

### 5) Submit a PR
- Use a clear PR title.
- Explain the problem and the fix.
- Include any relevant screenshots/logs for UI/extension changes.

## Development notes

### Embeddings / models
- Embedding model selection must use `config.GetEmbedModel()`.
- Summarization model selection must use `config.GetSummarizeModel()`.

### Avoid hardcoded module paths
- All Go imports should point to `github.com/Oridjinnn/Rewind`.

## License
By contributing, you agree that your contributions will be licensed under the MIT License.

