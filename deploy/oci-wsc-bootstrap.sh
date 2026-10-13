#!/usr/bin/env bash
# Installs the WSC game servers and a TLS-protected token bridge on a VM you control.
# Designed for Ubuntu 22.04/24.04 on an Oracle Cloud Always Free-eligible x86 VM.
set -Eeuo pipefail
export TERM=${TERM:-xterm}

: "${WSC_BRIDGE_DOMAIN:?Set WSC_BRIDGE_DOMAIN to your DNS name before running}"
: "${ACME_EMAIL:?Set ACME_EMAIL to an email for HTTPS certificate registration}"

if [[ "$(id -u)" -ne 0 ]]; then
  echo "Run this bootstrap as root." >&2
  exit 1
fi

apt-get update
DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates curl git openssl ufw docker.io docker-compose-v2
systemctl enable --now docker
docker version
docker compose version

PUBLIC_IP="$(curl -4fsSL https://api.ipify.org)"
[[ -n "$PUBLIC_IP" ]] || { echo "Could not determine public IPv4." >&2; exit 1; }

STACK=/opt/pretendo-wii-sports-club
if [[ ! -d "$STACK/.git" ]]; then
  git clone --depth 1 https://github.com/NVDEMU/pretendo-wii-sports-club.git "$STACK"
fi
cd "$STACK"

# Keep all secrets on this VM. The generated bridge token is shared only with the authorized
# Pretendo account-service operator over a private channel; never put it in GitHub or a public issue.
umask 077
cat > .env <<EOF
KERBEROS_PASSWORD=$(openssl rand -hex 32)
WSC_INTERNAL_API_TOKEN=$(openssl rand -hex 32)
SECURE_SERVER_LOCATION=$PUBLIC_IP
SECURE_SERVER_PORT=60015
NEX_AUTH_PUBLIC_HOST=$PUBLIC_IP
NEX_AUTH_PUBLIC_PORT=60014
WSC_BRIDGE_DOMAIN=$WSC_BRIDGE_DOMAIN
ACME_EMAIL=$ACME_EMAIL
EOF
chmod 0600 .env

# Open only the HTTPS certificate endpoints and the two WSC UDP service ports in the guest firewall.
# Add matching OCI VCN ingress rules separately; restrict SSH and DNS administration to your own IP.
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 60014/udp
ufw allow 60015/udp
ufw delete allow 8080/tcp >/dev/null 2>&1 || true
ufw --force enable

docker compose -f docker-compose.yml -f docker-compose.public.yml up --build -d
docker compose -f docker-compose.yml -f docker-compose.public.yml ps

echo
echo "WSC server host: $PUBLIC_IP"
echo "NEX authentication: UDP 60014"
echo "NEX secure/game server: UDP 60015"
echo "HTTPS token bridge: https://$WSC_BRIDGE_DOMAIN/internal/wsc/nex-token"
echo "Private bridge bearer token is stored in $STACK/.env; do not display or post that value publicly."
echo "TCP 8080 and MongoDB are not published to the Internet."
echo "This starts the game server. Pretendo's hosted account/service-discovery route still needs operator-side registration."
