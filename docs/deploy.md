# Deploying Måltiden

Production runs as one Fly.io app (`maltiden`, region `arn` / Stockholm): a
single Docker image with the Go API, which also serves the built Vue app, plus
a SQLite database on a Fly volume.

## Release flow

- Feature branches → PR into `dev` → CI (`.github/workflows/ci.yml`) and
  Playwright (`playwright.yml`) run on the PR.
- `dev` → PR into `main` is a production release. On every push to `main` that
  touches `backend/`, `frontend/`, `Dockerfile`, `fly.toml` or the deploy
  workflow, `.github/workflows/deploy.yml` runs `flyctl deploy --remote-only`
  and then checks `/health`.
- Manual or emergency deploy: run the **Deploy to Fly** workflow from the
  Actions tab (`workflow_dispatch`). It asks for a reason, which is logged.

Migrations in `backend/migrations/` run automatically at startup, in the order
listed in `internal/storage/sqlite/db.go`. There are no down-migrations, so
take a snapshot before deploying anything destructive (see [Backups](#backups)).

## Configuration

Plain settings live in `fly.toml` under `[env]` (`DATABASE_PATH`,
`FRONTEND_URL`). Everything sensitive is a Fly secret (`fly secrets set NAME=…`)
and must never be committed.

| Variable | Required | Purpose |
|---|---|---|
| `JWT_SECRET` | yes | HS256 signing key, at least 32 characters. The server refuses to start without it. Rotating it logs everyone out. |
| `ANTHROPIC_API_KEY` | no | Turns on the AI recipe parser (`POST /recipes/parse*`). Without it those routes aren't registered. |
| `GEMINI_API_KEY` (or `GOOGLE_API_KEY`) | no | Turns on the optional menu AI layer (wishes + weekday arrangement). Without it, menus are generated deterministically. |
| `RESEND_API_KEY`, `EMAIL_FROM` | no | Password-reset email. Without the key, the reset link is only written to the server log. |
| `SENTRY_DSN` | no | Backend error reporting. |
| `CORS_ORIGINS` | no | Comma-separated list of allowed origins. Production is same-origin, so it's only needed for split setups. |
| `AI_PARSER_DAILY_PER_HOUSEHOLD` / `AI_PARSER_DAILY_GLOBAL` | no | Daily cap on recipe-parser requests (default 20 per household / 200 total). `0` turns the feature off. |
| `AI_MENU_DAILY_PER_HOUSEHOLD` / `AI_MENU_DAILY_GLOBAL` | no | Daily cap on AI-assisted menu generations (default 30 per household / 1500 total). Over the cap, generation quietly falls back to the deterministic menu. |

The caps reset at 00:00 UTC and are kept in memory. See `pkg/aibudget`.

GitHub Actions uses the repository secret `FLY_API_TOKEN` and the repository
variable `VITE_SENTRY_DSN_FE`. The frontend DSN is baked into the JS bundle,
so it's public by design.

**Keep it to one machine.** SQLite on a Fly volume only works with one writer,
so do not `fly scale count` above 1. `auto_stop_machines` lets the machine sleep
when idle, and the first request after that has a cold start.

## Backups

### Current state

The only backup is Fly's automatic **volume snapshots** of `maltiden_data`: one
a day, kept for 5 days (checked Feb 2026). They are better than nothing, but:

- you can lose up to 24 hours of data;
- they are stored with the same provider as the live volume;
- restoring means creating a new volume from a snapshot and moving the machine
  onto it, which takes minutes and a careful hand.

Useful commands:

```sh
fly volumes list -a maltiden              # note the vol_… id
fly volumes snapshots list vol_REPLACE_ME
fly volumes snapshots create vol_REPLACE_ME   # manual snapshot before risky migrations
```

The privacy policy (`frontend/src/views/PrivacyView.vue`) tells users that
deleted data is gone from backups within five days. If backup retention
changes, update that page in the same PR.

### Recommended: Litestream to object storage

Not set up yet, because it needs bucket credentials added as Fly secrets.
[Litestream](https://litestream.io) streams the SQLite WAL to S3-compatible
storage every few seconds, so you can restore to within seconds of a failure,
and the copy lives outside the volume. The database already runs in WAL mode,
which Litestream requires.

1. Create a bucket. With Fly's Tigris integration, `fly storage create` sets
   `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_ENDPOINT_URL_S3`,
   `AWS_REGION` and `BUCKET_NAME` as secrets on the app. Any S3-compatible
   provider works if you set those secrets yourself.
2. Add `litestream.yml` to the image:

   ```yaml
   dbs:
     - path: /app/data/maltiden.db
       replicas:
         - type: s3
           bucket: ${BUCKET_NAME}
           path: maltiden
           endpoint: ${AWS_ENDPOINT_URL_S3}
           region: ${AWS_REGION}
           retention: 72h   # keep within the 5 days promised in the privacy policy
   ```

3. In the `Dockerfile` runtime stage, install the Litestream release binary and
   copy the config in. Change the entrypoint so Litestream supervises the server
   and forwards shutdown signals, which matters because Fly auto-stops the
   machine:

   ```dockerfile
   CMD ["litestream", "replicate", "-config", "/app/litestream.yml", "-exec", "./server"]
   ```

4. On start, restore when the volume is empty (a fresh volume or a new region):
   `litestream restore -if-db-not-exists -if-replica-exists -config /app/litestream.yml /app/data/maltiden.db`
   before `replicate`. A small shell entrypoint script is the usual way.
5. Test a restore into a scratch app before relying on it.

Check these steps against the Litestream docs for the version you install.
Config keys have changed between releases.
