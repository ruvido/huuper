# AGENTS

Repository guidance for contributors and automated agents.

## Purpose
Minimal guidelines to contribute to the project consistently, while keeping simplicity, clarity, and alignment with the repository stack.

## Principles
- **Less is more**: keep code simple, clear, and minimal.
- **DRY always**: avoid duplication of logic, data, and processes.
- **Official docs first**: follow PocketBase/Go/Svelte official conventions.
- Prefer state-of-the-art solutions that remain simple to operate in self-hosting.
- When in doubt, choose the simplest and clearest option with the fewest moving parts.

## Stack (Repository Constraints)
- Backend: PocketBase (Go framework) with custom APIs/hooks in Go.
- Frontend: Svelte with hash routing.
- Deploy: single Go binary.

## Operational Rules
- No custom code when a standard/reliable solution already exists.
- Do not make architectural or scope decisions without discussing first.
- Before implementing structural changes, propose options and wait for confirmation.
- Keep changes small and focused; avoid redundancy.
- Do not introduce opaque automation that is hard to maintain.
- Do not add new dependencies unless strictly necessary.

## Key Notes (Repo-Specific)
- Server-side logic must be in Go (no JS SDK on the server).
- Use `.env` to set default admin credentials.
- Frontend package manager is **bun only** (do not use npm).
- Keep the original routing and project tree.

## Frontend
- Use semantic CSS.
- Minimize the number of classes.
- Keep hash routing (`#/...`) as currently configured in the project.
- Follow the existing frontend folder structure in the repository (no structural refactors unless explicitly requested).

## Do / Don't
Do:
- Prefer existing helpers/components.
- Keep logic linear.
- Document only what is non-obvious.
- Propose options before structural changes.

Don't:
- Duplicate code.
- Add clever abstractions.
- Drift from official conventions.
- Introduce feature creep.
- Introduce fallback logic or automatic behavior not explicitly requested.

## Collaboration Constraint
- Do not introduce automatic behavior, inferred UX changes, or fallback logic unless explicitly requested.
- If a change is ambiguous, ask before implementing.

## When Unsure
Check official docs first:
- https://pocketbase.io/docs/
- https://pocketbase.io/docs/use-as-framework/
- https://svelte.dev/docs/svelte/
- https://svelte.dev/docs/kit/configuration#router

## Useful Commands
```bash
./pocketbase serve              # Start dev
cd frontend && bun install      # Install frontend deps
cd frontend && bun run build    # Build frontend
go build -o pocketbase main.go  # Build binary
```
