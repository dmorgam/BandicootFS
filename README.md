# BandicootFS

![BandicootFS](./assets/logo.png)

BandicootFS is a Go-based MCP server that turns your local filesystem into a
retrieval layer for AI models — no vector database required. It exposes a base
folder as MCP tools, letting any MCP-compatible client (Claude Desktop, IDE
integrations, custom agents) search and reason over your data on demand.
Lightweight, dependency-free, and runs locally.

> **Status: Work in Progress — Highly Incomplete**
>
> Early development. Most features are **not yet implemented**. Expect breaking
> changes and unstable APIs.

## Requirements

- Go 1.26+

## Build & run

```sh
go run . -root /path/to/docs            # stdio (Claude Desktop, Inspector)
go build -o bandicootfs . && ./bandicootfs -root /path/to/docs
```

HTTP transport:

```sh
go run . -transport http -port 8000 -root .
curl localhost:8000/healthz             # -> ok
```

### Flags

| Flag         | Default  | Description                          |
|--------------|----------|--------------------------------------|
| `-transport` | `stdio`  | `stdio` \| `http`                    |
| `-root`        | `.`     | base folder to index (tool default)  |
| `-topk`        | `5`     | default number of chunks per query   |
| `-chunk-lines` | `80`    | lines per chunk                      |
| `-overlap`     | `20`    | overlapping lines between chunks      |
| `-min-score`   | `0`     | drop chunks scoring at or below this |
| `-bm25-k1`     | `1.2`   | BM25 term-frequency saturation       |
| `-bm25-b`      | `0.75`  | BM25 length normalization (0..1)     |
| `-port`        | `8000`  | listen port (http only)              |
| `-path`        | `/mcp`  | mount path (http only)               |

## Tools

### `get_file_chunks`

Retrieve the chunks most related to a prompt within a base folder.

| Param         | Required | Description                                    |
|---------------|----------|------------------------------------------------|
| `prompt`      | yes      | natural-language query                         |
| `base_folder` | no       | folder to search; defaults to `-root`          |

Returns `chunks[]` of `{ text, source, start_line, score }`, ranked best first.

### `grep`

Regex search over indexable files. Paginated.

| Param         | Required | Description                                    |
|---------------|----------|------------------------------------------------|
| `pattern`     | yes      | regular expression                             |
| `base_folder` | no       | sub-folder under root; defaults to root        |
| `ignore_case` | no       | case-insensitive match                         |
| `limit`       | no       | max matches (default 50, max 200)              |
| `offset`      | no       | matches to skip                                |

### `list_files`

List indexable files. Paginated.

| Param         | Required | Description                                    |
|---------------|----------|------------------------------------------------|
| `base_folder` | no       | sub-folder under root; defaults to root        |
| `glob`        | no       | filename glob, e.g. `*.go`                     |
| `limit`       | no       | max files (default 50, max 200)                |
| `offset`      | no       | files to skip                                  |

**Pagination:** `grep` and `list_files` return `{ items[], total, next_offset }`.
`next_offset` is `-1` when the last page is reached; otherwise pass it back as
`offset` for the next page. This caps response size to protect the client's
context window.

**Sandbox:** `grep`/`list_files` resolve `base_folder` relative to `-root` and
reject any path escaping it (no `../` traversal).

## How retrieval works

No embedding model. Ranking is lexical, rebuilt per call:

```
walk(base_folder) -> split files into overlapping line windows
                  -> tokenize (lowercase, drop stopwords / short tokens)
                  -> BM25 score vs prompt -> drop <= min-score -> top-K
```

**BM25** normalizes for chunk length and saturates repeated terms, so large
chunks no longer drown short queries. Still lexical, not semantic (synonyms
won't match). Tunable via flags / `retriever` options: `ChunkLines` (80),
`Overlap` (20), `TopK` (5), `K1` (1.2), `B` (0.75), `MinScore` (0); plus `Exts`.
To go semantic later, swap the tokenizer/BM25 for a dense embedder + cosine.

## Project layout

```
main.go              entry: flags + transport selection
internal/
  transport/         stdio + streamable HTTP
  server/            builds *mcp.Server, registers tools
  tools/             MCP tool shells (thin adapters -> domain)
  retriever/         domain: TF-IDF + cosine retrieval (no MCP types)
```

Each tool is a thin shell in `tools/` that maps MCP input/output to a call into
the `retriever` domain package.

## Testing

```sh
go test ./...
go test -race ./...
```

- `internal/retriever` — unit tests: ranking + edge cases (empty prompt, bad folder).
- `internal/tools` — end-to-end over an in-memory MCP transport (client ↔ server,
  real `tools/call`).

Interactive:

```sh
npx @modelcontextprotocol/inspector go run . -root .
```

Claude Desktop (`claude_desktop_config.json`):

```json
{ "mcpServers": { "bandicoot": { "command": "go", "args": ["run", ".", "-root", "/path/to/docs"] } } }
```
