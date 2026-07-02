#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"

BFF="teamgramd/etc2/bff.yaml"
IP_FILE=".public_ip"
SECRET_FILE=".turn_secret"
ENV_FILE=".env"
PROFILE_FILE=".env_profile"

# --- Infrastructure profile (interactive) -------------------------------
# Picks which dependency stack to run. All three share the same core
# (kafka/etcd/redis/mysql/minio); they differ in memory limits and the
# logging/tracing/metrics services.
DEFAULT_PROFILE="default"
[[ -f "$PROFILE_FILE" ]] && DEFAULT_PROFILE="$(tr -d '[:space:]' < "$PROFILE_FILE")"

echo "Choose an infrastructure profile:"
echo "  1) minimal  - core only + strict memory limits        (~2-4 GB RAM)"
echo "  2) default  - core only, no memory limits             (~4-8 GB RAM)"
echo "  3) full     - core + logging / tracing / metrics      (~16 GB RAM)"
read -rp "Profile [${DEFAULT_PROFILE}]: " PROFILE_CHOICE
PROFILE_CHOICE="${PROFILE_CHOICE:-$DEFAULT_PROFILE}"
case "$PROFILE_CHOICE" in
  1|min|minimal)  ENV_PROFILE="min" ;;
  2|default|def)  ENV_PROFILE="default" ;;
  3|full)         ENV_PROFILE="full" ;;
  *) echo "[ERROR] invalid profile: $PROFILE_CHOICE"; exit 1 ;;
esac
ENV_COMPOSE="docker-compose-env-${ENV_PROFILE}.yaml"
echo "$ENV_PROFILE" > "$PROFILE_FILE"
echo "[cfg] infrastructure profile = ${ENV_PROFILE} (${ENV_COMPOSE})"

# --- Public address (interactive) ---------------------------------------
# Public IP/host that remote clients use to reach this server. Baked into the
# MTProto + VoIP/TURN config so chats AND calls work globally (not only LAN).
DEFAULT_IP=""
[[ -f "$IP_FILE" ]] && DEFAULT_IP="$(tr -d '[:space:]' < "$IP_FILE")"
[[ -z "$DEFAULT_IP" ]] && DEFAULT_IP="$(curl -fsS --max-time 3 https://api.ipify.org 2>/dev/null || true)"

if [[ -n "$DEFAULT_IP" ]]; then
  read -rp "Public server IP/host [${DEFAULT_IP}]: " PUBLIC_IP
else
  read -rp "Public server IP/host: " PUBLIC_IP
fi
PUBLIC_IP="${PUBLIC_IP:-$DEFAULT_IP}"
[[ -z "$PUBLIC_IP" ]] && { echo "[ERROR] public IP/host is required."; exit 1; }
echo "$PUBLIC_IP" > "$IP_FILE"

# --- TURN secret (generated once, reused) -------------------------------
TURN_SECRET=""
[[ -f "$SECRET_FILE" ]] && TURN_SECRET="$(tr -d '[:space:]' < "$SECRET_FILE")"
if [[ -z "$TURN_SECRET" ]]; then
  TURN_SECRET="$(openssl rand -hex 24 2>/dev/null || head -c 24 /dev/urandom | base64 | tr -dc 'a-zA-Z0-9')"
  echo "$TURN_SECRET" > "$SECRET_FILE"
fi

# --- compose env (consumed by the coturn service) -----------------------
{
  echo "PUBLIC_IP=${PUBLIC_IP}"
  echo "TURN_SECRET=${TURN_SECRET}"
} > "$ENV_FILE"

# --- bake public address + TURN secret into the server config -----------
# Keep a pristine copy of the tracked config and ALWAYS restore it on exit
# (success or failure), so the rendered public IP + TURN secret are never left
# in the working tree and can't be committed/pushed by accident. The image is
# built from the rendered file before the script exits, so restoring afterwards
# is safe.
BFF_PRISTINE="$(mktemp)"
cp "$BFF" "$BFF_PRISTINE"
restore_bff() {
  if [[ -f "$BFF_PRISTINE" ]]; then
    cp "$BFF_PRISTINE" "$BFF"
    rm -f "$BFF_PRISTINE"
    echo "[cfg] bff.yaml reverted to placeholders (secret not left on disk)."
  fi
}
trap restore_bff EXIT

sed -i -E "s|^([[:space:]]*Ip:[[:space:]]*).*$|\1${PUBLIC_IP}|" "$BFF"
sed -i -E "s|^([[:space:]]*Password:[[:space:]]*).*$|\1\"${TURN_SECRET}\"|" "$BFF"

# --- render coturn config from template (public IP + TURN secret) --------
mkdir -p coturn
sed -e "s|__PUBLIC_IP__|${PUBLIC_IP}|g" -e "s|__TURN_SECRET__|${TURN_SECRET}|g" \
    coturn/turnserver.conf.template > coturn/turnserver.conf
echo "[cfg] public address = ${PUBLIC_IP}; TURN relay configured."

echo
# When switching down from the full profile, remove leftover observability-only
# containers (safe: only these fixed names, never the core or app container).
if [[ "$ENV_PROFILE" != "full" ]]; then
  docker rm -f jaeger grafana prometheus kibana elasticsearch filebeat go-stash node-exporter >/dev/null 2>&1 || true
fi

echo "[1/3] docker compose -f ${ENV_COMPOSE} up -d  (infrastructure: ${ENV_PROFILE})"
docker compose -f "${ENV_COMPOSE}" up -d

echo
echo "[1.5/3] waiting for MySQL, then applying DB migrations (idempotent)"
for _i in $(seq 1 60); do
  docker exec mysql mysql -uteamgram -pteamgram teamgram -N -e "SELECT 1" >/dev/null 2>&1 && break
  sleep 2
done
if docker exec mysql mysql -uteamgram -pteamgram teamgram -N -e "SELECT 1" >/dev/null 2>&1; then
  docker exec mysql mysql -uteamgram -pteamgram teamgram -e "CREATE TABLE IF NOT EXISTS schema_migrations (name VARCHAR(128) NOT NULL PRIMARY KEY, applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);" >/dev/null 2>&1
  # Apply each migrate-*.sql in date order, only if not already recorded. --force tolerates
  # "already applied" duplicate-column/key errors on a pre-existing DB, so nothing breaks.
  for f in $(ls teamgramd/deploy/sql/migrate-*.sql 2>/dev/null | sort); do
    name=$(basename "$f")
    done=$(docker exec mysql mysql -uteamgram -pteamgram teamgram -N -e "SELECT COUNT(1) FROM schema_migrations WHERE name='${name}'" 2>/dev/null)
    if [ "$done" != "1" ]; then
      echo "  [migrate] ${name}"
      docker exec -i mysql mysql --force -uteamgram -pteamgram teamgram < "$f" >/dev/null 2>&1
      docker exec mysql mysql -uteamgram -pteamgram teamgram -e "INSERT IGNORE INTO schema_migrations (name) VALUES ('${name}');" >/dev/null 2>&1
    fi
  done
  echo "[OK] DB migrations applied."

# Sync ID generator counters from MySQL to Redis (prevents duplicate key issues)
if [ -f "./sync_idgen.sh" ]; then
  echo "[sync] Syncing ID generator counters (message_box_ngen, pts_updates_ngen)..."
  bash ./sync_idgen.sh
  # Start watchdog in background to auto-sync on drift
  echo "[sync] Starting ID generator watchdog (background)..."
  nohup bash ./sync_idgen.sh watchdog > ./sync_watchdog.log 2>&1 &
  WATCHDOG_PID=$!
  echo $WATCHDOG_PID > ./sync_watchdog.pid
  echo "[sync] Watchdog started (PID: $WATCHDOG_PID)"
else
  echo "[WARN] sync_idgen.sh not found, skipping ID generator sync"
fi
else
  echo "[WARN] MySQL not ready ~120s - skipping migrations"
fi

echo
echo "[2/3] docker compose up -d --build  (core server)"
# --build so the edited config (public address) is baked into the image.
docker compose up -d --build

echo
echo "[3/3] coturn (calls relay) — best-effort, never blocks the core server"
docker compose -f docker-compose-turn.yaml up -d \
  || echo "[WARN] coturn failed to start — calls relay unavailable; the core server is fine."

echo
echo "[OK] Server started (public address: ${PUBLIC_IP})."
echo "     Open these ports in the VPS PROVIDER firewall (and OS firewall):"
echo "       TCP 10443        - MTProto (login / chats / media)"
echo "       UDP+TCP 3478     - TURN/STUN control (calls)"
echo "       UDP 49160-49200  - TURN media relay (calls)"
