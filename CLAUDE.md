# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Sub2API is an **AI API gateway** that distributes/manages API quotas from AI product subscriptions. Users get platform-issued API Keys; the gateway handles auth, billing, load balancing, and forwards requests to upstream AI providers (Anthropic/Claude, OpenAI/Codex, Gemini, Antigravity, Bedrock). Go backend (Ent ORM + Gin) + Vue 3 frontend, on PostgreSQL + Redis.

Note: upstream is `Wei-Shaw/sub2api` (go module path is still `github.com/Wei-Shaw/sub2api`). This fork tracks it and is rebased onto upstream `main`.

## Commands

All Make targets are at the repo root (`Makefile`) and in `backend/Makefile`. There is no `datamanagement/` dir in this tree, so skip the `*-datamanagementd` targets.

```bash
# Build (root)
make build                 # backend + frontend
make build-backend         # -> backend/Makefile build, outputs backend/bin/server

# Backend (run from backend/)
go run ./cmd/server/                 # run server
go test -tags=unit ./...             # unit tests
go test -tags=integration ./...      # integration tests (needs Postgres + Redis)
go test -tags=e2e -v ./internal/integration/...   # e2e (or backend `make test-e2e`)
golangci-lint run ./...              # lint (golangci-lint v2.7 pinned in CI)
go generate ./ent                    # regenerate Ent code after schema edits
go generate ./cmd/server             # regenerate server-side generated code

# Run a single test
cd backend && go test -tags=unit -run TestName ./internal/service/

# Frontend (use pnpm, NOT npm — run from frontend/)
pnpm install                         # must commit pnpm-lock.yaml when deps change
pnpm dev
pnpm build
pnpm run lint:check
pnpm run typecheck
pnpm exec vitest run <spec>          # single spec; `make test-frontend-critical` runs the curated set

make secret-scan                     # python3 tools/secret_scan.py
```

CI requires Go (see `backend/go.mod` for the exact version) and `pnpm install --frozen-lockfile`. Workflows: `backend-ci.yml` (tests + lint), `security-scan.yml` (govulncheck + gosec + pnpm audit, weekly), `release.yml` (tags `v*` only).

## Architecture

### Layered backend (`backend/internal/`)

`handler` (HTTP/Gin) → `service` (business logic, ~290 files — the core of the system) → `repository` (data access over Ent) → `ent` (generated ORM). Cross-cutting: `middleware`, `config`, `domain` (constants/types), `model` (DTOs), `pkg`/`util`, `payment`, `setup` (wiring), `server` (router/http bootstrap).

Routes are registered in `internal/server/routes/` split by surface: `gateway.go` (the AI proxy endpoints), `admin.go`, `auth.go`, `user.go`, `payment.go`, `common.go`. HTTP bootstrap is `internal/server/http.go` + `router.go`.

### The gateway path (the heart of the product)

A request to a gateway endpoint flows: API-key auth → account scheduling/selection → upstream forwarding with failover → billing/usage recording.

- **Scheduling**: `openai_account_scheduler.go`, `scheduler_*.go` (cache/events/outbox/snapshot) pick an upstream account, with sticky sessions and concurrency limits.
- **Forwarding**: provider-specific gateway services — `openai_gateway_*.go`, `antigravity_gateway_service.go`, `gateway_forward_as_chat_completions.go`, `gateway_forward_as_responses.go`, plus `http_upstream_port.go`/`http_upstream_profile.go` and TLS fingerprint profiles (`ent/schema/tls_fingerprint_profile.go`). Token providers per platform: `claude_token_provider.go`, `openai_token_provider.go`, `gemini_token_provider.go`, `antigravity_token_provider.go`.
- **Failover**: `handler/failover_loop.go` retries across accounts on configured upstream status codes.
- **Billing**: `billing_service.go`, `billing_cache_*.go`, `gateway_billing_*.go` do token-level usage tracking and cost calc; pricing data in `backend/resources/model-pricing/`.

Account/model **mapping** matters: upstream model names are mapped to platform model names per account. Mis-mapping (esp. OpenAI/Codex accounts during bulk edits) silently makes accounts unschedulable — see DEV_GUIDE.md 坑 10.

### Data layer — Ent ORM

Schema is defined in `backend/ent/schema/*.go`; everything else under `backend/ent/` is **generated**. After editing a schema you MUST `go generate ./ent` and commit the generated output. Key entities: `account`, `account_group`, `api_key`, `user`, `usage_log`, `subscription_plan`, `payment_order`, `auth_identity` (multi-provider OAuth), `channel_monitor*`, `tls_fingerprint_profile`. SQL migrations live in `backend/migrations/`.

### Auth & identity

Multi-provider OAuth/OIDC: LinuxDo, WeChat, DingTalk, Email, generic OIDC (`handler/auth_*_oauth.go`, `service/auth_*`). Identities are modeled separately from users via `auth_identity` + `auth_identity_channel` with a pending-bind flow.

### Payments

Self-service top-up built in: EasyPay, Alipay, WeChat Pay, Stripe (`internal/payment/`, `service/payment_*`, `ent/schema/payment_*`). See `docs/PAYMENT.md`.

### Frontend (`frontend/src/`)

Vue 3 + pnpm. Standard layout: `api/`, `components/`, `views/`, `stores/`, `composables/`, `router/`, `types/`, `i18n/`, `utils/`. Curated critical Vitest specs are listed in the root Makefile (`FRONTEND_CRITICAL_VITEST`) — keep those green.

## Gotchas (from DEV_GUIDE.md — read it for the full list)

- **pnpm only.** Mixing npm/pnpm corrupts `node_modules`; always commit `pnpm-lock.yaml` or `--frozen-lockfile` CI fails.
- **Interface changes** require updating every `Stub`/`Mock` in `internal/` test files, or compilation breaks.
- **Ent schema changes** require `go generate ./ent` + committing generated code.
- **Adding a config option**: update `internal/config`, `deploy/config.example.yaml`, and any wiring in `internal/setup`.

## Deploy

Docker-first. `deploy/` holds `docker-compose*.yml`, `Dockerfile`, `install.sh`, `Caddyfile`, `config.example.yaml`. Root `Dockerfile` is the canonical image; `Dockerfile.goreleaser` + `.goreleaser*.yaml` drive releases. See `deploy/DOCKER.md`.
