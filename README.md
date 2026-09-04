# vibescript-lang.org

The Vibescript site. It is a Go app that imports the upstream Vibescript examples corpus, serves the `.vibe` source through the web UI, and executes every example in the browser through a Go-hosted Vibescript runtime.

## Run

```bash
just run
```

The server listens on `0.0.0.0:8080` by default. Override it with `HOST`, `PORT`, and `SHUTDOWN_TIMEOUT`.

## Deploy

Deployments are configured for Miren on Vultr. See [docs/deployment.md](docs/deployment.md).

## Current shape

- Serves `201` examples: `35` imported from `github.com/xipkit/vibescript`, plus `135` Rosetta Code tasks and `31` showcase programs.
- Exposes detail pages with real `.vibe` source.
- Runs every example through `/api/examples/{slug}/run`. `TestEveryExampleIsRunnable` enforces that each one defines a top-level `run` function.
- Serves a language reference at `/reference`, rendered from Markdown with goldmark.

## Test

```bash
go test ./...
```
