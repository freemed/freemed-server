#!/usr/bin/env bash
#
# backup-db.sh — Dump the freemed MySQL database, compress, and rotate old backups.
#
# Prerequisites:
#   - Docker Compose project "freemed-server" must be running.
#   - Container name "freemed-server-db-1" must be reachable.
#
# Output:
#   backups/freemed_YYYYMMDD_HHMMSS.sql.gz  (7 most-recent kept)

set -euo pipefail

# Resolve script directory so we can find backups/ relative to the project root.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BACKUP_DIR="${PROJECT_DIR}/backups"

CONTAINER="freemed-server-db-1"
DB_USER="freemed"
DB_PASS="freemed"
DB_NAME="freemed"
TIMESTAMP="$(date '+%Y%m%d_%H%M%S')"
BACKUP_FILE="${BACKUP_DIR}/freemed_${TIMESTAMP}.sql.gz"
KEEP_COUNT=7

# ----------------------------------------------------------------------
# 1. Ensure backups directory exists
# ----------------------------------------------------------------------
mkdir -p "${BACKUP_DIR}"

# ----------------------------------------------------------------------
# 2. Dump & compress
# ----------------------------------------------------------------------
echo "[$(date '+%H:%M:%S')] Starting backup → ${BACKUP_FILE}"

docker exec "${CONTAINER}" mysqldump \
    -u "${DB_USER}" \
    -p"${DB_PASS}" \
    --no-tablespaces \
    --single-transaction \
    --routines \
    --triggers \
    "${DB_NAME}" \
    | gzip > "${BACKUP_FILE}"

echo "[$(date '+%H:%M:%S')] Backup complete ($(du -h "${BACKUP_FILE}" | cut -f1))"

# ----------------------------------------------------------------------
# 3. Rotate — keep only the KEEP_COUNT most-recent .sql.gz files
# ----------------------------------------------------------------------
# Collect all compressed backups sorted by name (which sorts by timestamp),
# then delete everything beyond the keep count.
BACKUPS=($(ls -1t "${BACKUP_DIR}"/freemed_*.sql.gz 2>/dev/null))
if [ ${#BACKUPS[@]} -gt ${KEEP_COUNT} ]; then
    echo "[$(date '+%H:%M:%S')] Rotating: ${#BACKUPS[@]} backups exist, keeping ${KEEP_COUNT}"
    for ((i=${KEEP_COUNT}; i<${#BACKUPS[@]}; i++)); do
        echo "  Removing ${BACKUPS[$i]}"
        rm -f "${BACKUPS[$i]}"
    done
else
    echo "[$(date '+%H:%M:%S')] No rotation needed (${#BACKUPS[@]} backup(s) ≤ ${KEEP_COUNT})"
fi

echo "[$(date '+%H:%M:%S')] Done."
