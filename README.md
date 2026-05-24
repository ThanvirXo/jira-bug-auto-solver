# Jira Auto Bug Solver

Automatically fixes frontend bugs filed in Jira. When a bug ticket is created, this service classifies it using a self-hosted LLM — if it's a frontend bug, it triggers Claude Code to locate the root cause and stage a fix in your frontend repository. Backend bugs are silently ignored.

## How it works

```
Jira Bug Created
      │
      ▼
Webhook Receiver (Go)
      │
      ▼
Self-Hosted LLM Classifier
      │
      ├── backend → ignore
      │
      └── frontend
            │
            ▼
Generate bug.md
            │
            ▼
Create Isolated Git Worktree
            │
            ▼
Claude Code Agent
            │
            ▼
Generate + Stage Patch
            │
            ▼
Human Review & Merge
```

Claude Code does not commit or push — it only stages the changes. You review and merge manually.

## Prerequisites

- Go 1.21+
- [Claude Code CLI](https://claude.ai/code) installed and authenticated
- A vLLM-compatible self-hosted LLM endpoint
- Jira webhook configured to point at this service

## Environment variables

Create a `.env` file (for local dev) or set these in your environment:

| Variable | Description |
|---|---|
| `API_KEY` | Bearer token Jira must send in the `Authorization` header |
| `LLM_URL` | Base URL of your vLLM server, e.g. `http://localhost:8000/v1` |
| `VLLM_KEY` | API key for your vLLM server |
| `LLM_MODEL` | Model ID served by vLLM, e.g. `ai-model` |
| `FRONTEND_REPO_PATH` | Absolute path to your frontend repository on this machine |
| `FRONTEND_BUGS_DIR` | Directory where bug report `.md` files are written (must be accessible from `FRONTEND_REPO_PATH`) |
| `CLAUDE_PATH` | Path to the Claude Code CLI binary, e.g. `/usr/local/bin/claude` |
| `APP_STAGE` | Set to `production` to skip `.env` loading; omit or set to `dev` for local development |

## Running

```bash
go run .
```

The server starts on port `8080`.

## Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/health` | None | Health check |
| `POST` | `/jira/webhook` | Bearer token | Receives Jira webhook events |

## Jira webhook setup

1. In Jira, go to **Settings → System → WebHooks**
2. Create a new webhook pointing to `https://<your-host>/jira/webhook`
3. Set the `Authorization` header to `Bearer <your API_KEY>`
4. Select the **Issue Created** event (and optionally **Issue Updated**)

## Bug classification

The LLM classifier uses a structured decision framework to handle ambiguous reports (e.g. "nothing is showing", "data not coming") where the reporter doesn't know the technical root cause. Frontend bugs are routed to Claude Code; backend bugs are logged and dropped.

## Project structure

```
├── classifier/       LLM-based frontend/backend classifier
├── common/           Shared response types
├── handlers/         HTTP handlers (health check, webhook)
├── jira/             Jira webhook payload types (supports ADF and plain text descriptions)
├── middlewares/      Auth and request logging middleware
├── services/         Core business logic — webhook handling, Claude Code trigger
├── writer/           Writes bug reports as markdown files
├── app.go            App wiring (routes, middleware)
└── main.go           Entry point
```
