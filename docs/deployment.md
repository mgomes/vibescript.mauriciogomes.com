# Deployment

This site deploys as a Miren app on a Vultr Linux server.

## Miren App

The app config lives in `.miren/app.toml`.

- App name: `vibescript`
- Service: `web`
- Command: `/bin/app`
- Port: `8080`
- Scaling: one fixed instance

Miren detects the Go app from `go.mod`, builds the binary on the server, and starts it as `/bin/app`.

## First-Time Vultr Migration

Use a fresh Vultr VM for the Miren cutover. The previous deployment stack used Caddy, systemd, and direct SSH binary uploads; Miren should own ingress and TLS on the new server instead of competing with those services in place.

1. Create a Vultr Linux VM that meets Miren's server requirements.
2. Install the Miren CLI on the VM.
3. Run `sudo miren server install` on the VM.
4. Bind the new cluster from the local machine with `miren login` and `miren cluster add`.
5. Preview the app detection:

   ```bash
   miren deploy --analyze
   ```

6. Deploy without moving production traffic:

   ```bash
   miren deploy
   ```

7. Set a temporary route and verify the app:

   ```bash
   miren route set preview.vibescript.mauriciogomes.com vibescript
   miren logs -a vibescript
   curl -fsS https://preview.vibescript.mauriciogomes.com/
   ```

8. Point `vibescript.mauriciogomes.com` at the new Vultr IP.
9. Set the production route:

   ```bash
   miren route set vibescript.mauriciogomes.com vibescript
   ```

10. Keep the old VM online until the production route and rollback path have been verified.

## Routine Deploys

After the cluster is configured:

```bash
miren deploy
```

Use Miren history and rollback for operational recovery:

```bash
miren app history -a vibescript
miren rollback -a vibescript
```

## Vibescript Version Bumps

The most common deploy is tracking a new upstream `vibescript` release. The version is pinned in two places: `go.mod` and the `UpstreamVersion` constant in `internal/catalog/catalog.go`. That constant feeds the header badge, the homepage version chip, and the per-example "view source" links.

Replace `vX.Y.Z` with the target tag:

```bash
go get github.com/mgomes/vibescript@vX.Y.Z
go mod tidy
# edit internal/catalog/catalog.go: UpstreamVersion = "vX.Y.Z"
go build ./...
go test ./...
```

`TestAllExamplesCompileAndPassStaticChecks` is the real signal that the new release doesn't break any embedded example. `TestEveryExampleIsRunnable` then asserts that all of them still expose a top-level `def run`.

Spot-check locally:

```bash
go run .
curl -s http://localhost:8080/ | grep brand-version
curl -s http://localhost:8080/examples/strings-operations | grep -oE 'blob/v[0-9.]+/[^"]+' | head -1
```

Commit as two atomic changes on `master` (matching prior bumps):

1. `Bump vibescript to X.Y.Z` — `go.mod`, `go.sum`
2. `Show vibescript X.Y.Z in version badge and source links` — `internal/catalog/catalog.go`

Deploy and verify:

```bash
miren deploy
curl -s https://vibescript.mauriciogomes.com/healthz
curl -s https://vibescript.mauriciogomes.com/ | grep brand-version
```

If the new tag exposes a runtime regression that only shows up in production, `miren rollback -a vibescript` reverts; then bisect upstream from a clean state.

## Credentials

Do not commit provider credentials, root passwords, or API tokens. Rotate any credentials that were previously stored in local deployment scripts before relying on the Miren deployment path.
