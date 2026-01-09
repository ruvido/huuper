# Deploy Guide

Guida rapida per il deploy di Huuper con Docker.

## Quick Start

```bash
# 1. Configura variabili d'ambiente
cp .env.example .env
# Modifica .env con i tuoi valori

# 2. Build e avvia
./deploy.sh build
./deploy.sh up

# 3. Accedi all'app
open http://localhost:8090
```

## Prerequisiti

- Docker
- Docker Compose
- File `.env` configurato

## Configurazione .env

Crea un file `.env` nella root con:

```bash
# Admin credentials (per setup iniziale)
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=your-secure-password

# Telegram Bot
TELEGRAM_BOT_TOKEN=your-bot-token

# (Aggiungi altre variabili necessarie)
```

## Comandi Deploy Script

### Build e Deploy
```bash
./deploy.sh build      # Build Docker image
./deploy.sh up         # Avvia servizi in background
./deploy.sh down       # Ferma servizi
./deploy.sh restart    # Riavvia servizi
./deploy.sh rebuild    # Rebuild completo (no cache)
```

### Monitoring
```bash
./deploy.sh status     # Stato servizi
./deploy.sh logs       # Segui i logs (Ctrl+C per uscire)
```

### Backup & Restore
```bash
./deploy.sh backup                    # Backup pb_data in backups/YYYYMMDD_HHMMSS/
./deploy.sh restore backups/20260109  # Ripristina da backup
```

### Maintenance
```bash
./deploy.sh clean      # Rimuovi container, immagini e volumi
```

## Struttura Volumi

```
huuper/
├── pb_data/           # Database e uploads (persistente)
│   ├── data.db
│   └── storage/
├── backups/           # Backup directory (creata da ./deploy.sh backup)
└── docker-compose.yml
```

## Branch Strategy

### Production (Frozen v1.0)
```bash
git checkout production
./deploy.sh rebuild
```

### Development (Main)
```bash
git checkout main
# Sviluppo normale
go run . serve
```

## Deploy su Server Remoto

### 1. Clone repository sul server
```bash
git clone <repo-url>
cd huuper
git checkout production  # Per versione stabile
```

### 2. Configura .env
```bash
nano .env
# Imposta valori production
```

### 3. Deploy
```bash
./deploy.sh build
./deploy.sh up
```

### 4. Verifica
```bash
./deploy.sh status
./deploy.sh logs
```

## Troubleshooting

### Port già in uso
```bash
# Cambia porta in docker-compose.yml
ports:
  - "8091:8090"  # Usa 8091 invece di 8090
```

### Container non si avvia
```bash
./deploy.sh logs       # Vedi errori
./deploy.sh rebuild    # Rebuild completo
```

### Reset completo
```bash
./deploy.sh down
rm -rf pb_data/        # ⚠️  Cancella tutti i dati!
./deploy.sh up
```

## Health Check

Il container include un health check automatico:
- Endpoint: `http://localhost:8090/api/health`
- Intervallo: 30s
- Timeout: 3s
- Retries: 3

Verifica manualmente:
```bash
curl http://localhost:8090/api/health
```

## Backup Automatico

Per backup automatico giornaliero, aggiungi un cron job:
```bash
# Apri crontab
crontab -e

# Aggiungi (backup ogni giorno alle 3:00 AM)
0 3 * * * cd /path/to/huuper && ./deploy.sh backup
```

## Update da v1.0

Quando esci dal freeze e vuoi aggiornare:

```bash
# Sul server production
git fetch
git checkout production
git pull origin production
./deploy.sh rebuild
```

## Ports

- `8090`: HTTP API e frontend

## Network

I container usano la rete `huuper-network` (bridge mode).

## Logs

I logs sono accessibili via:
```bash
./deploy.sh logs                    # Follow mode
docker compose logs huuper          # One-shot
docker compose logs huuper --tail=100  # Ultime 100 righe
```
