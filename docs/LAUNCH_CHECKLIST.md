# Soft Launch Checklist — Måltiden

**Target:** 2–5 beta families (~10–20 users) on `maltiden.fly.dev`
**Date created:** 2026-02-25

---

## 1. Pre-Deploy Verification

### Database & Migrations

- [x] All migrations (001–010) apply cleanly on a fresh database
  - Dependency order verified: users → households → recipes → seed → menus → invites → indexes → more seeds → token version → recipe household_id
  - All use `IF NOT EXISTS` and `INSERT OR IGNORE` for idempotency
  - `ALTER TABLE` statements (006, 009, 010) add columns with safe defaults
- [x] Seed recipes load correctly — **20 recipes** (5 via 004, 15 via 008)
  - Pasta Carbonara, Kycklingwok, Tacos, Laxfilé, Köttfärssås, Pannkakor, Kycklinggryta, Ärtsoppa, Falukorv, Fiskpinnar, Korvstroganoff, Janssons frestelse, Pytt i panna, Vegetarisk pasta med pesto, Stekt fläsk, Köttbullar, Ugnsbakad torsk, Chili con carne, Tomatsoppa med ostmacka, Kyckling med currysås
- [x] New columns exist: `users.token_version` (DEFAULT 1), `recipes.household_id` (nullable)
- [x] Server creates DB, runs migrations, and seeds data on empty `DATABASE_PATH`
  - WAL mode, foreign keys, busy timeout (5s), connection pool (10 max) configured

### Build & CI

- [x] **Backend:** `go build` → `go test -race` (93 tests pass) → `go vet` — all clean
- [x] **Frontend:** `npm ci` → `type-check` → `lint` → `build` — all clean
  - Build output: ~291 kB JS app + ~247 kB JS vendor + ~142 kB CSS + ~210 kB images
  - Gzipped total: ~190 kB (JS + CSS)
- [x] Docker image builds successfully — **99.2 MB** final image size
  - Multi-stage: Node build → Go build → debian:bookworm-slim runtime
- [ ] Verify Docker container starts and serves SPA + API correctly:
  ```bash
  docker run -e JWT_SECRET=$(openssl rand -hex 32) -p 8080:8080 maltiden-test
  # Visit http://localhost:8080 — should show landing page
  # Visit http://localhost:8080/health — should return {"status":"ok","db":"connected"}
  ```

### Environment Variables (Fly.io)

Check with `fly secrets list`:

- [x] `JWT_SECRET` — set, deployed (**verify it's ≥32 chars** — code enforces this on startup)
- [x] `DATABASE_PATH` — set via `fly.toml` env: `/app/data/maltiden.db`
- [ ] `ANTHROPIC_API_KEY` — **NOT SET** — AI recipe parsing disabled without it
  ```bash
  fly secrets set ANTHROPIC_API_KEY=sk-ant-...
  ```
- [x] `CORS_ORIGINS` — not needed in production (SPA and API on same origin, frontend uses relative URLs)

**Not needed:**
- No `GIN_MODE` equivalent — Go stdlib `net/http` has no debug mode
- No debug flags to disable — structured JSON logging via `slog` is production-ready

---

## 2. Fly.io Infrastructure

### Volume

- [x] Volume exists: `vol_r635e3l1jj1x36nr`, 1 GB, region `arn`, encrypted
- [x] Mounted at `/app/data` (confirmed in `fly.toml`)
- [x] Daily snapshots active — **6 snapshots** available, 5-day retention
  - Latest: 18 hours ago (1.6 KiB – 34 MiB range)
- **Restore command:**
  ```bash
  # List snapshots
  fly volumes snapshots list vol_r635e3l1jj1x36nr

  # Restore from snapshot (creates new volume)
  fly volumes create maltiden_data --snapshot-id <SNAPSHOT_ID> --region arn --size 1

  # Then update machine to use new volume
  fly machine update <MACHINE_ID> --volume <NEW_VOLUME_ID>:/app/data
  ```
- **Manual backup:**
  ```bash
  fly ssh sftp get /data/maltiden.db ./backups/maltiden-$(date +%Y%m%d).db -a maltiden
  ```

### Machine & Scaling

- [x] Machine: `shared-cpu-1x`, **256 MB RAM** — sufficient for 10–20 users with SQLite
- [x] Auto-stop enabled (`auto_stop_machines = 'stop'`, `min_machines_running = 0`)
  - **Impact:** First request after idle period takes ~3–5 seconds (machine boot + Go startup + SQLite open + migrations check)
  - **Recommendation for soft launch:** Acceptable for beta. If complaints arise, set `min_machines_running = 1` (costs ~$1.94/month for shared-cpu-1x)
- [ ] Test cold start time: stop the machine, then time a request:
  ```bash
  fly machine stop <MACHINE_ID>
  time curl -s https://maltiden.fly.dev/health
  ```

### Health Check

- [x] `fly.toml` now includes health check (added in this session):
  ```toml
  [[http_service.checks]]
    grace_period = "10s"
    interval = "30s"
    method = "GET"
    timeout = "5s"
    path = "/health"
  ```
- [x] `GET /health` returns `{"status":"ok","db":"connected"}` (HTTP 200)
- [x] Health endpoint checks DB connectivity with 2-second timeout
- [x] No sensitive information exposed in health response

### TLS/HTTPS

- [x] HTTPS works on `maltiden.fly.dev` (HTTP/2, Fly.io-managed TLS)
- [x] HTTP redirects to HTTPS (301 Moved Permanently)
- [x] `force_https = true` in `fly.toml`
- [x] HSTS header set in production (`max-age=31536000; includeSubDomains`)
  - Triggered by `FLY_APP_NAME` env var presence (set automatically by Fly.io)

### Security Headers

- [x] CSP updated to allow Google Fonts (fixed in this session):
  - `style-src 'self' 'unsafe-inline' https://fonts.googleapis.com`
  - `font-src 'self' https://fonts.gstatic.com`
- [x] Security middleware moved to wrap entire handler including static files (fixed in this session)
  - Previously only API routes got security headers; now static files and SPA fallback do too
- [x] Full security header set:
  - `X-Content-Type-Options: nosniff`
  - `X-Frame-Options: DENY`
  - `X-XSS-Protection: 1; mode=block`
  - `Referrer-Policy: strict-origin-when-cross-origin`
  - `Permissions-Policy: camera=(), microphone=(), geolocation=()`
  - `Content-Security-Policy` (restrictive baseline)
  - `Strict-Transport-Security` (production only)

---

## 3. Monitoring & Observability

### Logging

- [x] Structured JSON logging via Go `log/slog` (machine-parseable)
- [x] Request ID middleware adds `X-Request-Id` header to all API responses
- **Commands:**
  ```bash
  # Live logs
  fly logs

  # Recent errors
  fly logs | grep -i error

  # Filter by request ID
  fly logs | grep "abc123"

  # Check startup/shutdown
  fly logs | grep -E "starting|stopping|shutdown"
  ```

### Alerting

- [ ] Consider free uptime monitoring (e.g., UptimeRobot free tier):
  - Monitor: `https://maltiden.fly.dev/health`
  - Interval: 5 minutes
  - Alert: email on downtime
- [x] Fly.io provides basic health alerts via the dashboard when health checks fail
  - Requires the health check config added above

### Database Monitoring

- [x] Current DB size with seed data: **~3–34 MiB** (based on snapshot sizes)
- **Check production DB size:**
  ```bash
  fly ssh console -C "ls -lh /app/data/maltiden.db"
  ```
- **Size guidelines:** SQLite handles GBs easily. With 20 users, expect <100 MB even after months of use. Monitor if approaching 500 MB.

---

## 4. User-Facing Readiness

### First User Experience

- [ ] **Smoke test on maltiden.fly.dev after deploy:**
  1. Visit landing page — clear description of what the app does?
  2. Click "Kom igång" → register with email/password
  3. Automatic household creation on registration
  4. Browse 20 seed recipes
  5. Generate weekly menu (5–7 days)
  6. View shopping list with categories and checkboxes
  7. Copy invite code → second user joins household
  8. Settings modal → feedback link visible
  9. Logout → session cleared
  - **Target:** Complete flow in under 2 minutes

### Error Handling in Production

- [x] **Claude API unavailable:** App logs warning on startup, parse endpoints not registered, app works fully without AI parsing
- [x] **SQLite corruption:** Restore from volume snapshot (see Section 2)
- [x] **Memory issues:** 256 MB is generous for Go + SQLite. Go's GC handles memory well. If OOM occurs, Fly.io restarts the machine automatically
- [x] **JWT expiry:** 401 interceptor sets `session_expired` flag, redirects to login with Swedish toast message

### Swedish Content

- [ ] Verify: all error toasts in Swedish
- [ ] Verify: all empty states in Swedish
- [ ] Verify: no English debug text or placeholder content visible
- [ ] Verify: landing page hero, features, CTA all in Swedish

### Feedback Channel

- [x] Feedback button added to SettingsModal (added in this session):
  - "Skicka feedback" button → opens `mailto:maltiden.app@gmail.com`
  - **TODO:** Create the `maltiden.app@gmail.com` mailbox, or update the email address
  - Alternative: replace with a Messenger group link if preferred

---

## 5. Known Limitations (share with beta families)

Communicate these to beta users upfront:

1. **Inget "Glömt lösenord" ännu** — kontakta oss via feedback-mejl om ni behöver byta lösenord
2. **Ingen offline-support** — appen kräver internet
3. **AI-receptparsning kräver bra formatering** — kopiera hela receptet med ingredienser och instruktioner
4. **Första laddningen kan ta 3–5 sekunder** — appen startar automatiskt vid första besöket efter inaktivitet
5. **Appen är i beta** — vi samlar aktivt feedback och fixar buggar löpande

---

## 6. Deploy Procedure

### Pre-deploy

```bash
# 1. Ensure all changes are committed on current branch
git status

# 2. Run full CI locally
cd backend && go build ./... && go test ./... -race && go vet ./...
cd ../frontend && npm ci && npm run type-check && npm run lint && npm run build

# 3. Build Docker image to verify
docker build -t maltiden-test .
```

### Deploy

```bash
# 4. Merge to main (rebase, ff-only)
git checkout dev
git merge --ff-only feat/mvp-sprint-complete  # or current branch
git checkout main
git merge --ff-only dev

# 5. Deploy to Fly.io
fly deploy

# 6. Verify health check passes
curl https://maltiden.fly.dev/health
# Expected: {"status":"ok","db":"connected"}

# 7. Check security headers are present
curl -sI https://maltiden.fly.dev/ | grep -i -E 'x-content-type|x-frame|strict-transport|content-security'

# 8. Set ANTHROPIC_API_KEY if AI parsing should work
fly secrets set ANTHROPIC_API_KEY=sk-ant-...
```

### Post-deploy smoke test

```bash
# 9. Manual smoke test (2 minutes):
#    - Visit https://maltiden.fly.dev
#    - Register a new account
#    - Browse recipes (should see 20)
#    - Generate a weekly menu
#    - View shopping list
#    - Open settings → verify feedback link
#    - Logout
```

### Monitor

```bash
# 10. Watch logs for 30 minutes
fly logs

# Check for errors
fly logs | grep -i error
```

---

## 7. Rollback Plan

### Rollback to previous version

```bash
# List recent deployments
fly releases

# Rollback to previous release
fly deploy --image maltiden:deployment-<PREVIOUS_HASH>

# Or redeploy from a specific git commit
git checkout <COMMIT_HASH>
fly deploy
```

### Restore database from snapshot

```bash
# 1. List available snapshots
fly volumes snapshots list vol_r635e3l1jj1x36nr

# 2. Create new volume from snapshot
fly volumes create maltiden_data --snapshot-id <SNAPSHOT_ID> --region arn --size 1

# 3. Stop the machine
fly machine stop <MACHINE_ID>

# 4. Detach old volume and attach new one
fly machine update <MACHINE_ID> --volume <NEW_VOLUME_ID>:/app/data

# 5. Start the machine
fly machine start <MACHINE_ID>

# 6. Verify health
curl https://maltiden.fly.dev/health
```

### Responsibility

- **David** monitors logs and handles backend issues for the first 48 hours
- **Jesper** handles frontend issues and user-reported UI bugs
- **Philip** coordinates with beta families and triages feedback

---

## Summary: What Was Fixed in This Session

| Issue | Fix | Files Changed |
|-------|-----|---------------|
| CSP blocked Google Fonts (broke all typography) | Added `fonts.googleapis.com` to `style-src`, `fonts.gstatic.com` to `font-src` | `backend/pkg/middleware/security.go` |
| Security headers missing from static files/SPA | Moved Security middleware to wrap entire handler in `main.go` | `backend/cmd/server/main.go`, `backend/internal/api/router.go` |
| No health check in fly.toml | Added `[[http_service.checks]]` section | `fly.toml` |
| No feedback channel in app | Added "Skicka feedback" button with email link in SettingsModal | `frontend/src/components/dashboard/SettingsModal.vue` |

## Summary: What Needs Manual Attention Before Deploy

| Item | Priority | Action |
|------|----------|--------|
| Set `ANTHROPIC_API_KEY` in Fly.io | Medium | `fly secrets set ANTHROPIC_API_KEY=sk-ant-...` — without it, AI recipe parsing is disabled |
| Create feedback email inbox | Medium | Create `maltiden.app@gmail.com` or update the email in SettingsModal |
| Philip's QA walkthrough | High | Full manual test of all user flows on production |
| Cold start time test | Low | Measure actual boot time after machine stop |
| Uptime monitoring setup | Low | Free UptimeRobot account hitting `/health` every 5 min |
| npm audit vulnerabilities | Low | 5 transitive dependency vulnerabilities (1 moderate, 4 high) — no exploitable paths in app code |
| Share known limitations with beta families | Medium | See Section 5 above |
