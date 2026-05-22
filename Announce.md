# Announcement (Release / Update)

Rewind now has a more consistent configuration layer, improved recall quality, and better reliability in model/embedding handling.

## What’s changed

- Embedding model is now controlled via `REWIND_EMBED_MODEL` (`config.GetEmbedModel()`), removing remaining hardcoded `nomic-embed-text` usage.
- Configuration is centralized in `internal/config/config.go` (chat/summarize/embed models + IDE port).
- Recall + IDE cross-referencing behavior was corrected and embedding request error handling improved.
- CI/GitHub release wiring is in place (Go build/test/vet + Goreleaser).

## Demo

To generate the demo GIF locally:
- `bash ./make_demo.sh`

Note: this script requires additional external tools (e.g. `agg`).

## Contributing

See [`CONTRIBUTING.md`](./CONTRIBUTING.md).

