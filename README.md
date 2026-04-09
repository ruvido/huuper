# Huuper

Self-hosted webapp to manage private Telegram/Discord groups.

## Features

- User authentication (login/signup)
- User profile management
- Group listing with invite links
- Admin interface for group management
- Mobile-first responsive design

## Tech Stack

- **Backend**: PocketBase (Go framework)
- **Frontend**: skeleton-first static frontend served from `frontend/site`
- **Database**: SQLite (via PocketBase)
- **Deploy**: Docker (release-based)

## Getting Started

### Prerequisites

- Go 1.21+
### Installation

1. Clone the repository:
```bash
git clone https://github.com/YOUR_USERNAME/huuper.git
cd huuper
```

2. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your admin credentials
```

3. Run the server:
```bash
go run ./backend serve --http=127.0.0.1:8000 --dir=./data --frontend-dev
```

The app will be available at `http://127.0.0.1:8000`
The frontend is rebuilt from `frontend/skeleton` to `frontend/site` and hot-reloaded in dev mode.

### Building for Production

```bash
# Build Go binary
mkdir -p extra/bin
go build -o extra/bin/huuper ./backend

# Run
./extra/bin/huuper serve --dir=./data
```

### Deploy

```bash
./deploy/rsync.sh
```

Per dettagli e override VPS vedi [`docs/DEPLOY.md`](/Users/ruvido/notes/1_projects/15-realmen-bot-webapp/huuper/docs/DEPLOY.md).

## Project Structure

```
.
├── backend/            # Go backend
│   ├── api/
│   ├── bot/
│   ├── internal/
│   ├── migrations/
│   └── main.go
├── frontend/           # Active frontend v2
│   ├── skeleton/
│   └── site/
└── data/               # PocketBase data dir
```

## Development

The project follows these principles:
- **Less is more**: Simple, clear code
- **Official docs first**: Always consult official documentation
- **Best practices only**: Follow official conventions
- **Zero redundancy**: Avoid complex and duplicated code

## Request Pipeline

The request pipeline is configured in PocketBase through `settings.requests_flow`.

- The configured step actions are `assign_group`, `assign_guardian`, `mentoring`, `group_approved`, `admin_approved`.
- The runtime API exposes explicit action names: `set_group`, `set_guardian`, `set_mentoring_done`, `set_group_approved`, `set_admin_approved`.
- Each request stores a snapshot of the active flow when it is created, so later settings changes do not rewrite requests already in progress.
- Notification routing and content are also driven by the same flow configuration plus PocketBase templates.

See [`docs/REQUESTS_NOTIFICATIONS.md`](docs/REQUESTS_NOTIFICATIONS.md) for the compact reference.

## Avatar upload test

You can quickly verify that the `users` collection accepts JPEG, PNG, WebP, and GIF avatars thanks to the fixtures in `extra/test/images/`. Each format has a 1:1 sample file that mirrors what the onboarding flow produces.

1. Start PocketBase locally (e.g. `./launch.sh`).
2. From the repo root run:

```bash
POCKETBASE_URL=http://127.0.0.1:8090 node scripts/test-avatar-uploads.mjs
```

   - The script posts four users (one per image type) with generated emails like `webp-avatar+<timestamp>@test.local`.
   - Override `POCKETBASE_URL` if your API runs elsewhere; `TEST_PASSWORD` customizes the generated password (`Test1234!` by default).
3. Confirm the new accounts under `users` inside PocketBase (the CLI prints record IDs).

## License

MIT
