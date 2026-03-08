# vibescript.mauriciogomes.com

The Vibescript website scaffold. It ships as a small Go web app built on `chi`, with an examples catalog shape that can grow into hundreds of runnable examples without changing the routing model.

## Run

```bash
go run .
```

The server listens on `:8080` by default. Override it with `HOST`, `PORT`, and `SHUTDOWN_TIMEOUT`.

## Test

```bash
go test ./...
```

