#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PIDS=()
DB_CONTAINER="localization-postgres"

cleanup() {
    echo ""
    echo "Shutting down..."
    if [ ${#PIDS[@]} -gt 0 ]; then
        for pid in "${PIDS[@]}"; do
            kill "$pid" 2>/dev/null || true
        done
    fi
    wait 2>/dev/null
    echo "Stopping PostgreSQL container..."
    docker stop "$DB_CONTAINER" 2>/dev/null || true
    echo "All processes stopped."
}

trap cleanup EXIT INT TERM

# Load .env if it exists
if [ -f "$SCRIPT_DIR/.env" ]; then
    echo "Loading .env..."
    set -a
    source "$SCRIPT_DIR/.env"
    set +a
fi

# Start Colima if not running
if ! colima status >/dev/null 2>&1; then
    echo "Starting Colima..."
    if ! colima start 2>&1; then
        echo "Colima failed to start. Cleaning up stale state and retrying..."
        colima delete -f 2>/dev/null || true
        colima start
    fi
fi

# Start PostgreSQL via docker run
echo "Starting PostgreSQL..."
if docker ps -q -f name="$DB_CONTAINER" | grep -q .; then
    echo "PostgreSQL container already running."
else
    docker rm -f "$DB_CONTAINER" 2>/dev/null || true
    docker volume create localization-pgdata >/dev/null 2>&1 || true
    docker run -d \
        --name "$DB_CONTAINER" \
        -e POSTGRES_USER=localization \
        -e POSTGRES_PASSWORD=localization \
        -e POSTGRES_DB=localization \
        -p 5432:5432 \
        -v localization-pgdata:/var/lib/postgresql/data \
        -v "$SCRIPT_DIR/backend/internal/database/migrations:/docker-entrypoint-initdb.d" \
        postgres:17 >/dev/null

    echo "Waiting for PostgreSQL to be ready..."
    for i in $(seq 1 30); do
        if docker exec "$DB_CONTAINER" pg_isready -U localization >/dev/null 2>&1; then
            echo "PostgreSQL is ready."
            break
        fi
        if [ "$i" -eq 30 ]; then
            echo "ERROR: PostgreSQL failed to start."
            exit 1
        fi
        sleep 1
    done
fi

# Start Go backend
echo "Starting backend on :8080..."
cd "$SCRIPT_DIR/backend"
go run . &
PIDS+=($!)

# Wait for backend to be ready before starting frontend
echo "Waiting for backend to be ready..."
for i in $(seq 1 30); do
    if curl -s http://localhost:8080/api/health >/dev/null 2>&1; then
        echo "Backend is ready."
        break
    fi
    if [ "$i" -eq 30 ]; then
        echo "ERROR: Backend failed to start."
        exit 1
    fi
    sleep 1
done

# Start React frontend
echo "Starting frontend on :5173..."
cd "$SCRIPT_DIR/frontend"
npm run dev &
PIDS+=($!)

echo ""
echo "================================"
echo "  Backend:  http://localhost:8080"
echo "  Frontend: http://localhost:5173"
echo "================================"
echo "Press Ctrl+C to stop all processes"
echo ""

wait
