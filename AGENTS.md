# AGENTS.md

Go 1.26+ proxy server providing OpenAI/Gemini/Claude/Codex compatible APIs with OAuth and round-robin load balancing.

## Repository
- GitHub: https://github.com/router-for-me/CLIProxyAPI

## Commands
```bash
gofmt -w . # Format (required after Go changes)
go build -o cli-proxy-api ./cmd/server # Build
go run ./cmd/server # Run dev server
go test ./... # Run all tests
go test -v -run TestName ./path/to/pkg # Run single test
go build -o test-output ./cmd/server && rm test-output # Verify compile (REQUIRED after changes)
```
- Common flags: `--config <path>`, `--tui`, `--standalone`, `--local-model`, `--no-browser`, `--oauth-callback-port <port>`

## Config
- Default config: `config.yaml` (template: `config.example.yaml`)
- `.env` is auto-loaded from the working directory
- Auth material defaults under `auths/`
- Storage backends: file-based default; optional Postgres/git/object store (`PGSTORE_*`, `GITSTORE_*`, `OBJECTSTORE_*`)

## Deploy
- Develop domain: `https://seacloud-cli-proxy-api.cloud.seaart.dev`
- Production domain: `https://seacloud-cli-proxy-api.cloud.seaart.ai`
- Verified on 2026-04-15:
  - `GET /health` returns `200 {"status":"ok"}`
  - `GET /healthz` returns `200 {"status":"ok"}`
  - `GET /v1/models` without API key returns `401 {"error":"Missing API key"}`
- Production verified on 2026-04-25:
  - `GET /health` returns `200 {"status":"ok"}`
  - `GET /v1/models` with API key returns `200` and includes `gpt-5.3-codex`
  - `POST /v1/chat/completions` with `gpt-5.3-codex` returns `200` and the expected completion text

```bash
curl -sS -i https://seacloud-cli-proxy-api.cloud.seaart.dev/health
curl -sS -i https://seacloud-cli-proxy-api.cloud.seaart.dev/healthz
curl -sS -i https://seacloud-cli-proxy-api.cloud.seaart.dev/v1/models
curl -sS -i https://seacloud-cli-proxy-api.cloud.seaart.ai/health
```

## Loki Logs
- Base URL: `http://loki-gateway.us-central1.ops.vtrix.dev`
- Main cluster auth: `-u "seacloud-develop-us-central1:BbcHQmdfbZLtcYUMkzlv16vM9wArTAdO"`
- Deployment namespace: `base-installation-develop`
- Container / app / service name: `seacloud-cli-proxy-api`
- Current deployment logs are visible in Loki with labels such as `namespace="base-installation-develop"` and `container="seacloud-cli-proxy-api"`

```bash
# List available containers on the main cluster
curl -s "http://loki-gateway.us-central1.ops.vtrix.dev/loki/api/v1/label/container/values" \
  -u "seacloud-develop-us-central1:BbcHQmdfbZLtcYUMkzlv16vM9wArTAdO"

# Query recent logs for this service
curl -s -G "http://loki-gateway.us-central1.ops.vtrix.dev/loki/api/v1/query_range" \
  -u "seacloud-develop-us-central1:BbcHQmdfbZLtcYUMkzlv16vM9wArTAdO" \
  --data-urlencode 'query={namespace="base-installation-develop",container="seacloud-cli-proxy-api"}' \
  --data-urlencode 'limit=100' \
  --data-urlencode 'direction=backward'

# Query health-check logs only
curl -s -G "http://loki-gateway.us-central1.ops.vtrix.dev/loki/api/v1/query_range" \
  -u "seacloud-develop-us-central1:BbcHQmdfbZLtcYUMkzlv16vM9wArTAdO" \
  --data-urlencode 'query={namespace="base-installation-develop",container="seacloud-cli-proxy-api"} |= "/health"' \
  --data-urlencode 'limit=100' \
  --data-urlencode 'direction=backward'

# Query errors only
curl -s -G "http://loki-gateway.us-central1.ops.vtrix.dev/loki/api/v1/query_range" \
  -u "seacloud-develop-us-central1:BbcHQmdfbZLtcYUMkzlv16vM9wArTAdO" \
  --data-urlencode 'query={namespace="base-installation-develop",container="seacloud-cli-proxy-api"} |= "error"' \
  --data-urlencode 'limit=100' \
  --data-urlencode 'direction=backward'
```

## Architecture
- `cmd/server/` — Server entrypoint
- `internal/api/` — Gin HTTP API (routes, middleware, modules)
- `internal/api/modules/amp/` — Amp integration (Amp-style routes + reverse proxy)
- `internal/thinking/` — Main thinking/reasoning pipeline. `ApplyThinking()` (apply.go) parses suffixes (`suffix.go`, suffix overrides body), normalizes config to canonical `ThinkingConfig` (`types.go`), normalizes and validates centrally (`validate.go`/`convert.go`), then applies provider-specific output via `ProviderApplier`. Do not break this "canonical representation → per-provider translation" architecture.
- `internal/runtime/executor/` — Per-provider runtime executors (incl. Codex WebSocket)
- `internal/translator/` — Provider protocol translators (and shared `common`)
- `internal/registry/` — Model registry + remote updater (`StartModelsUpdater`); `--local-model` disables remote updates
- `internal/store/` — Storage implementations and secret resolution
- `internal/managementasset/` — Config snapshots and management assets
- `internal/cache/` — Request signature caching
- `internal/watcher/` — Config hot-reload and watchers
- `internal/wsrelay/` — WebSocket relay sessions
- `internal/usage/` — Usage and token accounting
- `internal/tui/` — Bubbletea terminal UI (`--tui`, `--standalone`)
- `sdk/cliproxy/` — Embeddable SDK entry (service/builder/watchers/pipeline)
- `test/` — Cross-module integration tests

## Code Conventions
- Keep changes small and simple (KISS)
- Comments in English only
- If editing code that already contains non-English comments, translate them to English (don’t add new non-English comments)
- For user-visible strings, keep the existing language used in that file/area
- New Markdown docs should be in English unless the file is explicitly language-specific (e.g. `README_CN.md`)
- As a rule, do not make standalone changes to `internal/translator/`. You may modify it only as part of broader changes elsewhere.
- If a task requires changing only `internal/translator/`, run `gh repo view --json viewerPermission -q .viewerPermission` to confirm you have `WRITE`, `MAINTAIN`, or `ADMIN`. If you do, you may proceed; otherwise, file a GitHub issue including the goal, rationale, and the intended implementation code, then stop further work.
- `internal/runtime/executor/` should contain executors and their unit tests only. Place any helper/supporting files under `internal/runtime/executor/helps/`.
- Follow `gofmt`; keep imports goimports-style; wrap errors with context where helpful
- Do not use `log.Fatal`/`log.Fatalf` (terminates the process); prefer returning errors and logging via logrus
- Shadowed variables: use method suffix (`errStart := server.Start()`)
- Wrap defer errors: `defer func() { if err := f.Close(); err != nil { log.Errorf(...) } }()`
- Use logrus structured logging; avoid leaking secrets/tokens in logs
- Avoid panics in HTTP handlers; prefer logged errors and meaningful HTTP status codes
- Timeouts are allowed only during credential acquisition; after an upstream connection is established, do not set timeouts for any subsequent network behavior. Intentional exceptions that must remain allowed are the Codex websocket liveness deadlines in `internal/runtime/executor/codex_websockets_executor.go`, the wsrelay session deadlines in `internal/wsrelay/session.go`, the management APICall timeout in `internal/api/handlers/management/api_tools.go`, and the `cmd/fetch_antigravity_models` utility timeouts
