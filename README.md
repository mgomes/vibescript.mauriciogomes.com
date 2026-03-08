# vibescript.mauriciogomes.com

The Vibescript site. It is a Go `chi` app that imports the upstream Vibescript examples corpus, serves the `.vibe` source through the web UI, and executes the runnable subset in the browser through a Go-hosted Vibescript runtime.

## Run

```bash
go run .
```

The server listens on `:8080` by default. Override it with `HOST`, `PORT`, and `SHUTDOWN_TIMEOUT`.

## Current shape

- Imports `34` upstream examples from `github.com/mgomes/vibescript`.
- Exposes detail pages with real `.vibe` source.
- Runs the examples that define a top-level `run` function through `/api/examples/{slug}/run`.

## Test

```bash
go test ./...
```
