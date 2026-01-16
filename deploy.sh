#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

ensure_frontend_deps() {
    if [ ! -d "frontend/node_modules" ] || [ ! -d "frontend/node_modules/marked" ]; then
        info "Installing frontend dependencies..."
        if [ -f "frontend/package-lock.json" ]; then
            (cd frontend && npm ci)
        else
            (cd frontend && npm install)
        fi
    fi
}

# Check .env file exists
if [ ! -f .env ]; then
    error ".env file not found! Please create it from .env.example"
fi

# Parse command
case "${1:-}" in
    build)
        ensure_frontend_deps
        info "Building frontend..."
        cd frontend && npm run build && cd ..
        info "Building Docker image..."
        docker compose build
        info "Build completed successfully!"
        ;;

    up)
        ensure_frontend_deps
        info "Building frontend..."
        cd frontend && npm run build && cd ..
        info "Starting services..."
        docker compose up -d
        info "Services started!"
        info "Access the app at http://localhost:8090"
        ;;

    down)
        info "Stopping services..."
        docker compose down
        info "Services stopped!"
        ;;

    restart)
        info "Restarting services..."
        docker compose restart
        info "Services restarted!"
        ;;

    logs)
        docker compose logs -f huuper
        ;;

    rebuild)
        info "Rebuilding and restarting..."
        docker compose down
        ensure_frontend_deps
        info "Building frontend..."
        cd frontend && npm run build && cd ..
        docker compose build --no-cache
        docker compose up -d
        info "Rebuild completed!"
        ;;

    backup)
        info "Stopping services for safe backup..."
        docker compose down

        BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"
        mkdir -p "$BACKUP_DIR"

        info "Creating backup in $BACKUP_DIR..."

        # Safe backup: stop service first to ensure consistency
        if [ -d "pb_data" ]; then
            cp -r pb_data "$BACKUP_DIR/"
            info "Backup completed: $BACKUP_DIR"
        else
            error "pb_data directory not found!"
        fi

        info "Restarting services..."
        docker compose up -d
        info "Services restarted!"
        ;;

    restore)
        if [ -z "$2" ]; then
            error "Usage: ./deploy.sh restore <backup_directory>"
        fi
        if [ ! -d "$2/pb_data" ]; then
            error "Backup directory not found or invalid: $2/pb_data"
        fi

        warn "This will overwrite current pb_data. Continue? (y/N)"
        read -r response
        if [[ "$response" =~ ^[Yy]$ ]]; then
            info "Stopping services..."
            docker compose down

            # Safety: move current pb_data to restore directory
            if [ -d "pb_data" ]; then
                RESTORE_BACKUP="restore/pb_data.$(date +%Y%m%d_%H%M%S)"
                mkdir -p restore
                info "Moving current pb_data to: $RESTORE_BACKUP"
                mv pb_data "$RESTORE_BACKUP"
            fi

            info "Restoring from $2..."
            cp -r "$2/pb_data" .

            info "Starting services..."
            docker compose up -d
            info "Restore completed!"
            info "Previous data saved in: $RESTORE_BACKUP"
        else
            info "Restore cancelled"
        fi
        ;;

    clean)
        warn "This will remove all containers, images, and volumes. Continue? (y/N)"
        read -r response
        if [[ "$response" =~ ^[Yy]$ ]]; then
            info "Cleaning up..."
            docker compose down -v
            docker rmi huuper-huuper 2>/dev/null || true
            info "Cleanup completed!"
        else
            info "Cleanup cancelled"
        fi
        ;;

    status)
        docker compose ps
        ;;

    *)
        echo "Huuper Deploy Script"
        echo ""
        echo "Usage: ./deploy.sh [command]"
        echo ""
        echo "Commands:"
        echo "  build           Build Docker image"
        echo "  up              Start services"
        echo "  down            Stop services"
        echo "  restart         Restart services"
        echo "  logs            Show logs (follow mode)"
        echo "  rebuild         Rebuild from scratch and restart"
        echo "  backup          Backup pb_data directory"
        echo "  restore <dir>   Restore from backup directory"
        echo "  clean           Remove all containers, images, and volumes"
        echo "  status          Show service status"
        echo ""
        ;;
esac
