# Wii Sports Club Network Revival

A community project exploring replacement online services for **Wii Sports Club on Wii U**.

> **Status: work in progress.** This repository is not affiliated with Nintendo or Pretendo Network. Online play is not yet a turnkey service: endpoint discovery, account integration, and DNS routing must be configured and verified for each supported path.

## Goals

- Recreate the authentication and secure game-server services required by Wii Sports Club.
- Keep gameplay services independent of the DNS provider selected by the player.
- Document compatible connection paths for Pretendo SSSL DNS and WiiLink DNS.
- Support regional clubs and the game's separate service requirements without assuming one region or sport is enough.
- Make builds reproducible and deployments configurable.

## DNS compatibility: Pretendo and WiiLink

The project aims to support either connection path. The game backend does **not** need to know which public DNS provider a player selected; it needs to advertise reachable server endpoints, and DNS must resolve those endpoints correctly.

| Connection path | Documented DNS address | Important detail |
| --- | --- | --- |
| Pretendo SSSL | `88.198.140.154` | Pretendo's Wii U guide documents this as its SSSL resolver. |
| WiiLink DNS | `167.235.229.36` | WiiLink documents this resolver and says some lookups are forwarded to Pretendo Network. |

These addresses are upstream DNS services, not the IP address of this project's game server. This repository cannot add records to Pretendo's or WiiLink's public DNS infrastructure. To redirect original Wii Sports Club service hostnames to a self-hosted replacement, the relevant hostname mappings must be available through the selected resolver, or the player must use an explicitly configured DNS relay/override that forwards other lookups to the chosen upstream.

On Wii U, the secondary DNS server is a fallback, not a way to merge Pretendo and WiiLink responses. Choose one upstream profile for a given DNS relay. Do not advertise the two public DNS addresses as if both were used simultaneously.

See [DNS compatibility and deployment](docs/DNS-COMPATIBILITY.md).

## Server components

- `wsc-authentication/` — NEX authentication service.
- `wsc-secure/` — NEX secure/game service, including matchmaking and game data handlers.
- `docker-compose.yml` — local deployment scaffold with MongoDB.
- `.github/workflows/ci.yml` — compile checks for both Go modules.

The server components are based on the GPL-3.0-licensed ReviveTendo Wii Sports Club server source published by [BrewtendoNetwork/Wii-Sports-Club](https://github.com/BrewtendoNetwork/Wii-Sports-Club). Attribution and the original license are retained. This repository is a separate community-maintained distribution, not an official Pretendo release.

## Quick start

1. Install Docker and Docker Compose.
2. Copy `.env.example` to `.env`.
3. Set `KERBEROS_PASSWORD` to the same strong secret for both NEX services.
4. Set `SECURE_SERVER_LOCATION` to a public hostname or address that resolves to this deployment for players using **either** DNS path.
5. Start the services with `docker compose up --build`.

This boots the server processes and MongoDB. It does **not** automatically provision Pretendo Network IDs or add Wii Sports Club hostname mappings to public DNS. Account-store integration and correct endpoint routing are required before clients can log in.

## Development status

- [x] Add the initial authentication and secure-server codebase with upstream attribution.
- [x] Add a shared deployment configuration and reproducible build checks.
- [x] Document the two DNS paths and their routing limitations.
- [ ] Verify the actual Wii Sports Club hostnames and service discovery flow for each DNS setup.
- [ ] Integrate account provisioning/authentication with the intended account backend.
- [ ] Add captured-packet regression tests for login, club data, rankings, matchmaking, and each sport.
- [ ] Validate end-to-end play on real Wii U hardware and document tested regions/sports.

## Upstream references

- [Pretendo Wii U connection guide](https://pretendo.network/docs/install/wiiu)
- [WiiLink DNS guide](https://wiilink.ca/guide/dns/)
- [Wii Sports Club server source used as a baseline](https://github.com/BrewtendoNetwork/Wii-Sports-Club)
