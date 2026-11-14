# pinkman

Pinkman is a reusable backend router for AI execution targets.

It keeps routing decisions separate from orchestration. A caller supplies execution requirements; Pinkman discovers compatible targets, filters them, ranks them deterministically, and invokes the selected backend through an adapter.

## Responsibilities

- discover only targets that are actually available through configured backends;
- filter targets by backend, model and required capabilities;
- rank compatible targets deterministically;
- execute an explicitly selected target;
- keep provider-specific HTTP and CLI details behind adapters.

Pinkman does **not** decide how many models to run, whether they should run in parallel, whether a reviewer is required, or when an agent loop should stop. Those decisions belong to the orchestrating runtime.

## Initial transports

- OpenAI-compatible HTTP: discovery through `GET /v1/models`, execution through `POST /v1/responses`, Bearer API-key authentication.
- OpenCode CLI: discovery through `opencode models`, execution through `opencode run --standalone --format json` with an optional `--model` argument.

## Package layout

```text
.
├── adapters/
│   ├── openai/
│   └── opencode/
├── catalog.go
├── resolver.go
└── types.go
```

## Design constraints

- No silent fallbacks.
- Discovery is read-only.
- Repeated discovery over an unchanged environment is idempotent.
- Only discovered and compatible targets are returned to callers.
- Routing stays deterministic for the same request and target catalog.
