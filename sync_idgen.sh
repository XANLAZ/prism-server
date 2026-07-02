#!/bin/bash
# Sync idgen Redis counters from MySQL on startup
# Runs after migrations, before core server starts

set -euo pipefail

log_info() {
    echo -e "\033[0;34m[INFO]\033[0m $(date '+%Y-%m-%d %H:%M:%S') - $*"
}

log_success() {
    echo -e "\033[0;32m[SUCCESS]\033[0m $(date '+%Y-%m-%d %H:%M:%S') - $*"
}

log_error() {
    echo -e "\033[0;31m[ERROR]\033[0m $(date '+%Y-%m-%d %H:%M:%S') - $*" >&2
}

# Wait for Redis
log_info "Waiting for Redis..."
for i in $(seq 1 30); do
    if docker exec redis redis-cli ping >/dev/null 2>&1; then
        log_success "Redis is ready"
        break
    fi
    sleep 1
done

# Get all user_ids that have data in messages or user_pts_updates
log_info "Fetching user IDs from MySQL..."

# Sync message_box_ngen counters
log_info "Syncing message_box_ngen counters..."
USER_IDS=$(docker exec mysql mysql -uteamgram -pteamgram teamgram -N -e "SELECT DISTINCT user_id FROM messages;")
for USER_ID in $USER_IDS; do
    MAX_BOX=$(docker exec mysql mysql -uteamgram -pteamgram teamgram -N -e "SELECT MAX(user_message_box_id) FROM messages WHERE user_id=$USER_ID;")
    if [[ -n "$MAX_BOX" && "$MAX_BOX" != "NULL" ]]; then
        NEXT_ID=$((MAX_BOX + 1))
        docker exec redis redis-cli SET "message_box_ngen_${USER_ID}" "$NEXT_ID" >/dev/null
        log_success "  user_id=$USER_ID: message_box_ngen = $NEXT_ID (was $MAX_BOX)"
    fi
done

# Sync pts_updates_ngen counters
log_info "Syncing pts_updates_ngen counters..."
USER_IDS=$(docker exec mysql mysql -uteamgram -pteamgram teamgram -N -e "SELECT DISTINCT user_id FROM user_pts_updates;")
for USER_ID in $USER_IDS; do
    MAX_PTS=$(docker exec mysql mysql -uteamgram -pteamgram teamgram -N -e "SELECT MAX(pts) FROM user_pts_updates WHERE user_id=$USER_ID;")
    if [[ -n "$MAX_PTS" && "$MAX_PTS" != "NULL" ]]; then
        NEXT_ID=$((MAX_PTS + 1))
        docker exec redis redis-cli SET "pts_updates_ngen_${USER_ID}" "$NEXT_ID" >/dev/null
        log_success "  user_id=$USER_ID: pts_updates_ngen = $NEXT_ID (was $MAX_PTS)"
    fi
done

log_success "ID generator sync completed"

# ============================================================
# WATCHDOG MODE: continuous sync (run in background)
# Usage: bash /root/prism-server/sync_idgen.sh watchdog &
# ============================================================
if [[ "${1:-}" == "watchdog" ]]; then
    log_info "Starting watchdog mode (sync every 300s)..."
    while true; do
        sleep 300
        log_info "[watchdog] Periodic sync check..."
        
        # Quick check: compare Redis vs MySQL for a few users
        USER_IDS=$(docker exec mysql mysql -uteamgram -pteamgram teamgram -N -e "SELECT DISTINCT user_id FROM messages LIMIT 10;")
        for USER_ID in $USER_IDS; do
            REDIS_VAL=$(docker exec redis redis-cli GET "message_box_ngen_${USER_ID}" 2>/dev/null || echo "0")
            MYSQL_VAL=$(docker exec mysql mysql -uteamgram -pteamgram teamgram -N -e "SELECT MAX(user_message_box_id) FROM messages WHERE user_id=$USER_ID;" 2>/dev/null)
            MYSQL_VAL=${MYSQL_VAL:-0}
            EXPECTED=$((MYSQL_VAL + 1))
            
            if [[ "$REDIS_VAL" != "$EXPECTED" ]]; then
                log_error "[watchdog] DESYNC detected for user $USER_ID: Redis=$REDIS_VAL, MySQL=$MYSQL_VAL, expected=$EXPECTED"
                docker exec redis redis-cli SET "message_box_ngen_${USER_ID}" "$EXPECTED" >/dev/null
                log_success "[watchdog] Fixed: message_box_ngen_${USER_ID} = $EXPECTED"
            fi
            
            # Check pts
            REDIS_PTS=$(docker exec redis redis-cli GET "pts_updates_ngen_${USER_ID}" 2>/dev/null || echo "0")
            MYSQL_PTS=$(docker exec mysql mysql -uteamgram -pteamgram teamgram -N -e "SELECT MAX(pts) FROM user_pts_updates WHERE user_id=$USER_ID;" 2>/dev/null)
            MYSQL_PTS=${MYSQL_PTS:-0}
            EXPECTED_PTS=$((MYSQL_PTS + 1))
            
            if [[ "$REDIS_PTS" != "$EXPECTED_PTS" ]]; then
                log_error "[watchdog] DESYNC pts for user $USER_ID: Redis=$REDIS_PTS, MySQL=$MYSQL_PTS, expected=$EXPECTED_PTS"
                docker exec redis redis-cli SET "pts_updates_ngen_${USER_ID}" "$EXPECTED_PTS" >/dev/null
                log_success "[watchdog] Fixed: pts_updates_ngen_${USER_ID} = $EXPECTED_PTS"
            fi
        done
    done
fi
