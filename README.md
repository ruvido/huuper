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
- **Frontend**: Svelte with hash routing
- **Database**: SQLite (via PocketBase)
- **Deploy**: Single binary

## Getting Started

### Prerequisites

- Go 1.21+
- Node.js 18+
- npm

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

3. Install frontend dependencies:
```bash
cd frontend
npm install
```

4. Build the frontend:
```bash
npm run build
```

5. Run the server:
```bash
cd ..
go run . serve --http=127.0.0.1:8000
```

The app will be available at `http://127.0.0.1:8000`

### Building for Production

```bash
# Build frontend
cd frontend
npm run build

# Build Go binary
cd ..
go build -o huuper main.go

# Run
./huuper serve
```

## Project Structure

```
.
├── frontend/           # Svelte frontend
│   ├── src/
│   │   ├── components/ # Reusable components
│   │   ├── lib/        # Utilities and stores
│   │   └── pages/      # Page components
│   └── package.json
├── migrations/         # Database migrations
├── main.go            # Go entry point
└── pb_public/         # Static files (generated)
```

## Development

The project follows these principles:
- **Less is more**: Simple, clear code
- **Official docs first**: Always consult official documentation
- **Best practices only**: Follow official conventions
- **Zero redundancy**: Avoid complex and duplicated code

## Avatar upload test

You can quickly verify that the `users` collection accepts JPEG, PNG, WebP, and GIF avatars thanks to the fixtures in `test/images/`. Each format has a 1:1 sample file that mirrors what the onboarding flow produces.

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
