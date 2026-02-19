#!/usr/bin/env bash
set -euo pipefail

# migrate-to-prod.sh
#
# Migrates ModulaCMS from the old modula-postgres compose stack to the new
# dedicated modula (prod) compose stack. Copies Docker volumes to new names
# and starts the prod stack.
#
# Run from the modulacms repo root on the production server:
#   bash deploy/docker/migrate-to-prod.sh
#
# What this does:
#   1. Stops the old modula-postgres stack
#   2. Copies each volume to its new prod-prefixed name
#   3. Starts the new modula prod stack
#   4. Runs a health check
#   5. Optionally removes old volumes

OLD_COMPOSE="deploy/docker/docker-compose.postgres.yml"
NEW_COMPOSE="deploy/docker/docker-compose.prod.yml"

# Old volume name -> New volume name
# Docker compose prepends the project name to volume names.
# Old project: modula-postgres, New project: modula
declare -A VOLUME_MAP=(
    ["modula-postgres_postgres_cms_data"]="modula_prod_cms_data"
    ["modula-postgres_postgres_cms_ssh"]="modula_prod_cms_ssh"
    ["modula-postgres_postgres_cms_backups"]="modula_prod_cms_backups"
    ["modula-postgres_postgres_db_data"]="modula_prod_db_data"
    ["modula-postgres_postgres_minio_data"]="modula_prod_minio_data"
)

log() { echo >&2 "[migrate] $*"; }
err() { echo >&2 "[migrate] ERROR: $*"; }

# --- Pre-flight checks ---

if [ ! -f "$OLD_COMPOSE" ]; then
    err "Old compose file not found: $OLD_COMPOSE"
    err "Run this script from the modulacms repo root."
    exit 1
fi

if [ ! -f "$NEW_COMPOSE" ]; then
    err "New compose file not found: $NEW_COMPOSE"
    exit 1
fi

log "Checking old volumes exist..."
missing=0
for old_vol in "${!VOLUME_MAP[@]}"; do
    if ! docker volume inspect "$old_vol" > /dev/null 2>&1; then
        err "Volume not found: $old_vol"
        missing=1
    fi
done

if [ "$missing" -eq 1 ]; then
    err "Some old volumes are missing. Check volume names with: docker volume ls"
    err "If your volumes have different names, edit the VOLUME_MAP in this script."
    exit 1
fi

log "All old volumes found."

# Check no new volumes already exist (avoid accidental overwrite)
for new_vol in "${VOLUME_MAP[@]}"; do
    if docker volume inspect "$new_vol" > /dev/null 2>&1; then
        err "New volume already exists: $new_vol"
        err "Remove it first or this is a re-run. Aborting to avoid data loss."
        exit 1
    fi
done

# --- Confirmation ---

log ""
log "This will:"
log "  1. Stop the old modula-postgres stack"
log "  2. Copy ${#VOLUME_MAP[@]} volumes to new names"
log "  3. Start the new modula prod stack"
log ""
log "Volume mapping:"
for old_vol in "${!VOLUME_MAP[@]}"; do
    log "  $old_vol -> ${VOLUME_MAP[$old_vol]}"
done
log ""

read -r -p "[migrate] Proceed? (y/N) " confirm
if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
    log "Aborted."
    exit 0
fi

# --- Step 1: Stop old stack ---

log "Stopping old stack..."
docker compose -f "$OLD_COMPOSE" down
log "Old stack stopped."

# --- Step 2: Copy volumes ---

log "Copying volumes..."
for old_vol in "${!VOLUME_MAP[@]}"; do
    new_vol="${VOLUME_MAP[$old_vol]}"
    log "  $old_vol -> $new_vol"
    docker volume create "$new_vol" > /dev/null
    docker run --rm \
        -v "${old_vol}:/src:ro" \
        -v "${new_vol}:/dst" \
        alpine sh -c 'cp -a /src/. /dst/'
done
log "All volumes copied."

# --- Step 3: Start new stack ---

log "Building and starting new prod stack..."
DOCKER_BUILDKIT=1 docker compose -f "$NEW_COMPOSE" up -d --build
log "New stack started."

# --- Step 4: Health check ---

log "Waiting for health check (up to 30s)..."
healthy=0
for i in $(seq 1 15); do
    sleep 2
    http_code=$(curl -Lkso /dev/null -w '%{http_code}' https://localhost/api/v1/health 2>/dev/null || echo "000")
    if [ "$http_code" != "000" ] && [ "$http_code" != "502" ] && [ "$http_code" != "503" ]; then
        log "Health check passed (HTTP $http_code)"
        healthy=1
        break
    fi
    log "  Attempt $i/15: HTTP $http_code"
done

if [ "$healthy" -eq 0 ]; then
    err "Health check failed after 30s."
    err "Check logs: docker compose -f $NEW_COMPOSE logs modula"
    err "Old volumes are preserved. To rollback:"
    err "  docker compose -f $NEW_COMPOSE down"
    err "  docker compose -f $OLD_COMPOSE up -d"
    exit 1
fi

# --- Step 5: Cleanup old volumes ---

log ""
log "Migration successful. Old volumes can now be removed:"
log ""
for old_vol in "${!VOLUME_MAP[@]}"; do
    log "  docker volume rm $old_vol"
done
log ""
read -r -p "[migrate] Remove old volumes now? (y/N) " cleanup
if [ "$cleanup" = "y" ] || [ "$cleanup" = "Y" ]; then
    for old_vol in "${!VOLUME_MAP[@]}"; do
        docker volume rm "$old_vol"
        log "  Removed $old_vol"
    done
    log "Old volumes removed."
else
    log "Old volumes kept. Remove them manually when ready."
fi

log "Migration complete."
