# Deploy Guide

Deploy is deterministic: build locally, sync artifacts to VPS, start container on VPS.

## Defaults

- `VPS_HOST=fiber`
- `VPS_PATH=/home/ruvido/dev/huuper`
- `SERVICE_NAME=huuper`
- `BIN_NAME=huuper`

You can override any of them per command.

## One Command Deploy

```bash
./deploy.sh
```

What it does:

1. Stops remote container (`docker compose down`).
2. Builds frontend locally (`frontend/npm ci && npm run build`).
3. Builds Linux binary locally (`huuper`).
4. Rsyncs `docker-compose.yml`, `Dockerfile`, `huuper`, `migrations/`, `pb_public/` to VPS.
5. Starts remote container with `docker compose up -d --build --force-recreate`.

## Required on VPS

Inside `/home/ruvido/dev/huuper` on VPS:

- `.env` must exist and be valid.
- `pb_data/` must exist (or will be created by docker volume mount flow).

## Override Example

```bash
VPS_HOST=ruvido@fiber VPS_PATH=/home/ruvido/dev/huuper ./deploy.sh
```

## Troubleshooting

### Check remote status

```bash
ssh fiber "cd /home/ruvido/dev/huuper && docker compose ps"
```

### Check logs

```bash
ssh fiber "cd /home/ruvido/dev/huuper && docker compose logs --tail=200 huuper"
```

### Force clean rebuild

```bash
ssh fiber "cd /home/ruvido/dev/huuper && docker compose down"
./deploy.sh
```
