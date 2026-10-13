# Public WSC host with Pretendo accounts

This setup runs **only the Wii Sports Club game servers**. It does not create a replacement PNID/account network. Players keep using their existing Pretendo Network accounts.

## What it enables

- Public WSC NEX authentication on UDP 60014 and secure/game service on UDP 60015.
- Persistent MongoDB and datastore storage.
- An HTTPS token-bridge endpoint for the account-server integration. The underlying TCP 8080 helper and MongoDB port remain private.
- A generated WSC-specific NEX password per authenticated Pretendo NEX PID. The account route returns the WSC password, registers the matching NEX token with the Pretendo account service, and keeps the same PID identity. This avoids needing to copy Pretendo's account database.

## 1. Create an eligible free VM

Use a provider console to create Ubuntu 22.04 or 24.04 on an Always Free-eligible VM. Oracle Cloud currently lists both AMD VM.Standard.E2.1.Micro (1 GB RAM) and Arm VM.Standard.A1.Flex allocations as Always Free resources; A1 with its available free OCPU/RAM allocation is preferable for building the Go containers. Capacity varies and account verification may be required. **Choose only a shape and boot volume that explicitly show no charge**; do not select a paid shape or leave a trial running after its terms expire. An ephemeral GitHub Actions runner cannot host permanent UDP game traffic.

## 2. Create a free HTTPS DNS name

Register a free DuckDNS subdomain at https://www.duckdns.org/ and keep its update token private. The bootstrap updates that subdomain to the VM's public IPv4 and waits for DNS resolution before starting Caddy, because Caddy needs to obtain an HTTPS certificate. A domain is used only to secure the account-to-game-server bridge; the Wii U's NEX endpoint still advertises the VM's public IPv4.

## 3. Open cloud-network ingress

In the provider's VCN/security-list/firewall panel, add these rules:

| Source | Destination protocol/port | Reason |
| --- | --- | --- |
| Your public IP/32 | TCP 22 | SSH/diagnostics only if needed |
| 0.0.0.0/0 | TCP 80, TCP 443 | HTTPS certificate and token bridge |
| 0.0.0.0/0 | UDP 60014 | WSC NEX authentication |
| 0.0.0.0/0 | UDP 60015 | WSC secure/game service |

Do **not** expose TCP 8080, MongoDB 27017, or any database/admin interface. The bridge is HTTPS with a long bearer secret, and only its single path is reverse-proxied.

## 4. Run the bootstrap through the provider's browser console

Inspect [the bootstrap script](../deploy/oci-wsc-bootstrap.sh), [the Compose overlay](../docker-compose.public.yml), and [the Caddy route](../deploy/Caddyfile) before running them. In the VM's cloud-init/initialization-script field, paste the following after replacing the two uppercase placeholders:

```bash
#!/bin/bash
set -eu
export WSC_BRIDGE_DOMAIN="YOUR-SUBDOMAIN.duckdns.org"
export ACME_EMAIL="YOUR-EMAIL-ADDRESS"
export WSC_DUCKDNS_TOKEN="YOUR-DUCKDNS-UPDATE-TOKEN"
apt-get update
apt-get install -y curl
curl -fsSL https://raw.githubusercontent.com/NVDEMU/pretendo-wii-sports-club/main/deploy/oci-wsc-bootstrap.sh -o /root/oci-wsc-bootstrap.sh
bash /root/oci-wsc-bootstrap.sh
```

The script installs Docker, updates your DuckDNS A record to the instance IP, creates random passwords and a bridge bearer token, writes a root-readable `.env`, builds the two Go server images sequentially to reduce memory pressure, and starts the WSC containers. The DuckDNS token is used only to update the DNS record and is not written into the WSC `.env` or public logs. The account bridge's public HTTPS path is:

```text
https://YOUR-SUBDOMAIN.duckdns.org/internal/wsc/nex-token
```

After first boot, confirm that Caddy obtained a valid TLS certificate and that `docker compose -f docker-compose.yml -f docker-compose.public.yml ps` shows the services running. The VM's public IP and port are the WSC endpoints; the HTTPS hostname is for the bridge.

## 5. Connect the public Pretendo account route

The repository contains a WSC-only patch for Pretendo's account route. It recognizes game-server ID `1012F100`, validates a WSC title ID, uses the already-authenticated Pretendo PNID and its existing NEX/game PID, issues and registers a NEX token, and asks the WSC bridge for the response XML.

An operator with access to Pretendo's hosted account service must deploy/approve this WSC-only route or register an equivalent route in its service-discovery configuration. Provide them the following **privately**:

- The route patch: [provider-route.patch](../integrations/pretendo-account/provider-route.patch)
- The bridge URL above
- The generated `WSC_INTERNAL_API_TOKEN` value from `/opt/pretendo-wii-sports-club/.env`, shared only through a private channel

They would configure the account-service environment with `WSC_INTERNAL_API_URL` set to that HTTPS URL and `WSC_INTERNAL_API_TOKEN` set to the same secret. Do not publish the token in GitHub issues, screenshots, console logs, or chat. The code in this repository cannot write to Pretendo's hosted production server, and public DNS alone does not register the game server.

## 6. Hardware validation

Using a Wii U with a Pretendo PNID, first confirm its account-service response advertises the VM IP and UDP 60014. Then verify that the WSC auth response advertises UDP 60015. Test login before matchmaking, records, and matches. If the hosted route has not been deployed, the game will continue to use Pretendo's existing response and this WSC server will not be reached.

## Status honesty

Successful image builds and container startup are not the same as a Wii U login. This guide becomes end-to-end only after the public Pretendo account route has been registered and a real console test succeeds. No public Pretendo or WiiLink DNS records are changed by this repository.
