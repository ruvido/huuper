# AGENTS

Repository guidance for contributors and automated agents.

## Core Principles
- **Less is more**: keep code simple, clear, and minimal.
- **DRY**: avoid duplication; reuse existing helpers/components.
- **Official docs first**: follow PocketBase/Go/Svelte docs and conventions.
- **Best practices only**: prefer standard patterns over cleverness.

## Stack
- Backend: PocketBase (Go framework) with custom APIs/hooks in Go.
- Frontend: Svelte with hash routing.
- Deploy: single Go binary.

## Key Notes
- Server-side logic must be in Go (no JS SDK on the server).
- Use `.env` to set default admin credentials.
- Keep changes small and focused; avoid redundancy.

## Do / Don't
Do: prefer existing helpers, keep logic linear, document only what is non-obvious.
Don't: duplicate code, add clever abstractions, drift from official conventions.

## When Unsure
Check official docs first:
- https://pocketbase.io/docs/
- https://pocketbase.io/docs/use-as-framework/
- https://svelte.dev/docs/svelte/
- https://svelte.dev/docs/kit/configuration#router

## Useful Commands
```bash
./pocketbase serve              # Start dev
cd frontend && npm run build    # Build frontend
go build -o pocketbase main.go  # Build binary
```
