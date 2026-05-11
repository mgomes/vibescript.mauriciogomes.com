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

## Credentials

Do not commit provider credentials, root passwords, or API tokens. Rotate any credentials that were previously stored in local deployment scripts before relying on the Miren deployment path.
