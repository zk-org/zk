# Embeddings

Use the `[embedding]` section to enable semantic indexing and natural-language search (`--match-strategy nl`).

```toml
[embedding]
enabled = true
provider = "local" # openai, googleai, local
endpoint = "http://127.0.0.1:11434/v1" # required for local
model = "nomic-embed-text"
```

## Quick start

1. Enable embeddings in `.zk/config.toml`:

```toml
[embedding]
enabled = true
provider = "local"
endpoint = "http://127.0.0.1:11434/v1"
model = "nomic-embed-text"
```

2. Re-index notes to generate chunk embeddings:

```sh
./zk index -f
```

3. Run a natural-language query:

```sh
./zk list -Mn -m "notes about rust concurrency and deadlocks"
```

## Options

- `enabled` (bool): turn semantic indexing and search on.
- `provider` (string): one of `openai`, `googleai`, `local`.
- `model` (string): embedding model name.
- `endpoint` (string): API base URL, required for `local`.
- `api-key-env` (string): env var name containing API key.
  - `openai` default: `ZK_EMBEDDING_OPENAI_API_KEY`
  - `googleai` default: `ZK_EMBEDDING_GOOGLE_API_KEY`
- `dimensions` (int): expected vector dimensions (`0` = auto detect).
- `chunk-size` (int): chunk size for note splitting, default `800`.
- `chunk-overlap` (int): overlap between chunks, default `200`.
- `max-chunks-per-note` (int): cap per note, default `128`.
- `batch-size` (int): texts per embedding request, default `32`.
- `query-top-k` (int): nearest chunks used for ranking, default `40`.
- `vector-weight` (float): hybrid ranking weight for vectors, default `0.7`.

## Local provider

`provider = "local"` targets OpenAI-compatible local embedding servers (for example Ollama, vLLM, LM Studio, llama.cpp-compatible servers).
