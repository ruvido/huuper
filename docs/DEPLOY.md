# Deploy Guide (Docker + Releases)

Deploy is deterministic and release-based:
- build artifacts locally
- upload to `releases/<release_id>` on VPS
- atomically switch `current` symlink
- restart Docker service

## Defaults

- `VPS_HOST=fiber`
- `VPS_PATH=/home/ruvido/apps/huuper`
- `SERVICE_NAME=huuper`
- `APP_HOST_PORT=8090`
- `TARGET_GOARCH=amd64`

## Remote Layout

Under `$VPS_PATH`:

- `deploy/` Docker files (`docker-compose.yml`, `Dockerfile`)
- `releases/<release_id>/` immutable artifacts (`bin`, `backend/migrations`, `frontend/site`)
- `current -> releases/<release_id>` active release
- `shared/data/` persistent PocketBase data
- `shared/.env` runtime environment file (required)

## One Command Deploy

```bash
./deploy/rsync.sh
```

What it does:

1. Validates `frontend/site`.
2. Builds Linux binary locally.
3. Uploads artifacts to a new release folder on VPS.
4. Updates `current` symlink atomically.
5. Runs `docker compose up -d --build --force-recreate`.
6. Waits for `/api/health` success.

## Required on VPS

Create `.env` locally in project root before deploy.
`deploy/rsync.sh` copies it automatically to `$VPS_PATH/shared/.env`.

```bash
cp .env.example .env
# edit .env values
```

`shared/data/` is created automatically if missing.

## Rollback

List releases:

```bash
ssh fiber "ls -1 /home/ruvido/apps/huuper/releases"
```

Rollback to a release:

```bash
./scripts/ops/rollback.sh <release_id>
```

## Override Example

```bash
VPS_HOST=ruvido@fiber VPS_PATH=/home/ruvido/apps/huuper ./deploy/rsync.sh
```

## Troubleshooting

Status:

```bash
ssh fiber "cd /home/ruvido/apps/huuper/deploy && docker compose -f docker-compose.yml ps"
```

Logs:

```bash
ssh fiber "cd /home/ruvido/apps/huuper/deploy && docker compose -f docker-compose.yml logs --tail=200 huuper"
```
