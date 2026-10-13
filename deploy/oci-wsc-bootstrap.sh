#!/usr/bin/env bash
# Optional bootstrap for a public WSC game host. No custom domain or HTTPS bridge required.
set -Eeuo pipefail
LOG_FILE=/var/log/wsc-bootstrap.log
exec > >(tee -a "$LOG_FILE") 2>&1

if [[ "$(id -u)" -ne 0 ]]; then
  echo "Run as root or through cloud-init." >&2
  exit 1
fi
: "${WSC_NEX_PASSWORD_SECRET:?Set WSC_NEX_PASSWORD_SECRET to the shared value configured for the account route}"

apt-get update
DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates curl git openssl ufw docker.io docker-compose-v2
systemctl enable --now docker
PUBLIC_IP="$(curl -4fsSL https://api.ipify.org)"
[[ -n "$PUBLIC_IP" ]] || { echo "Unable to determine public IPv4" >&2; exit 1; }

STACK=/opt/pretendo-wii-sports-club
if [[ ! -d "$STACK/.git" ]]; then
  git clone --depth 1 https://github.com/NVDEMU/pretendo-wii-sports-club.git "$STACK"
fi
cd "$STACK"

umask 077
if [[ ! -f .env ]]; then
  cat > .env <<EOF
KERBEROS_PASSWORD=$(openssl rand -hex 32)
WSC_NEX_PASSWORD_SECRET=$WSC_NEX_PASSWORD_SECRET
SECURE_SERVER_LOCATION=$PUBLIC_IP
SECURE_SERVER_PORT=60015
MONGO_URI=mongodb://mongo:27017/
NEX_ACCOUNTS_DATABASE=wsc
NEX_ACCOUNTS_COLLECTION=nexaccounts
EOF
  chmod 0600 .env
fi

ufw allow 22/tcp
ufw allow 60014/udp
ufw allow 60015/udp
ufw --force enable
docker compose -f docker-compose.cloud.yml up -d
docker compose -f docker-compose.cloud.yml ps

echo "WSC containers started on $PUBLIC_IP"
echo "NEX auth UDP 60014; secure/game UDP 60015"
echo "No domain or public token bridge required."
echo "The authorized account route must advertise $PUBLIC_IP:60014 and use the same password secret."
