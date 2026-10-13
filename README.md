# Wii Sports Club Network Revival

A community project exploring replacement online services for **Wii Sports Club on Wii U**.

> **Status: work in progress.** This repository is not affiliated with Nintendo, Pretendo Network, or WiiLink. The services compile and container images build, but end-to-end Wii U login and online play have not yet been verified.

## Goals

- Recreate the authentication and secure game-server services required by Wii Sports Club.
- Keep the game backend independent of the DNS provider selected by the player.
- Test the Pretendo SSSL DNS and WiiLink DNS connection paths separately.
- Support the game's regional clubs, ranking data, and matchmaking behavior.
- Make builds reproducible and deployment configuration explicit.

## DNS compatibility: Pretendo and WiiLink

The **server software is resolver-neutral**: it does not need to know whether a player uses Pretendo or WiiLink DNS. Both client paths must, however, reach the correct service-discovery and server endpoints.

| Connection path | Documented DNS address | What is verified |
| --- | --- | --- |
| Pretendo SSSL DNS | `88.198.140.154` | Listed by Pretendo's Wii U setup guide. |
| WiiLink DNS | `167.235.229.36` | Listed by WiiLink's DNS setup guide. |

These are DNS resolver addresses, not the address of this project's game server. Public resolver setup alone does not guarantee that the Wii Sports Club service-discovery flow advertises this project. The required hostname mappings and account-server response must be verified for each path; this repository cannot edit Pretendo's or WiiLink's public DNS records.

Do not configure these two resolver IPs as primary and secondary and expect records to merge: the secondary DNS is a fallback. See [DNS compatibility](docs/DNS-COMPATIBILITY.md) and the more important [Wii U service-discovery requirement](docs/SERVICE-DISCOVERY.md).

## Server components

- `wsc-authentication/` — NEX authentication service.
- `wsc-secure/` — NEX secure/game service, including matchmaking and game data handlers.
- `wsc-authentication/cmd/account-admin/` — command-line tool for provisioning a NEX password record for an existing PID.
- `POST /internal/wsc/nex-token` on private TCP 8080 — optional authenticated helper for a trusted account proxy; not a public Wii U endpoint.
- `docker-compose.yml` — local deployment scaffold with MongoDB.
- `config/wii-sports-club-endpoints.json` — machine-readable endpoint metadata with source URLs and confidence notes.
- `docs/ENDPOINT-REFERENCE.md` — sourced game-server ID, access key, account API route, historical Nintendo endpoint, and community ports.
- `docs/DEPLOYMENT.md` — Linux VPS setup, firewall, private account-service integration, and hardware verification sequence.
- `scripts/smoke_test_internal_bridge.py` — end-to-end check of the internal token bridge against MongoDB.
- `docs/SERVICE-DISCOVERY.md` — describes the Wii U account/service-discovery requirement.
- `integrations/pretendo-account/provider-route.patch` — optional WSC-only bridge patch for a self-hosted or explicitly authorized Pretendo Account Server; see its [integration guide](integrations/pretendo-account/README.md).
- `scripts/check_dns_routing.py` — queries a selected resolver for a verified hostname; can check the expected IPv4 result.
- `.github/workflows/ci.yml` — Go builds, endpoint/DNS unit tests, Compose validation, and container builds.
- `.github/workflows/dns-preflight.yml` — manually query a verified hostname through both public DNS resolvers.

The server components are based on the GPL-3.0-licensed Wii Sports Club server code published by [BrewtendoNetwork/Wii-Sports-Club](https://github.com/BrewtendoNetwork/Wii-Sports-Club). The upstream license and attribution are retained. This is a separate community-maintained project, not an official Pretendo release.

## Quick start: build the services

1. Install Docker and Docker Compose.
2. Copy `.env.example` to `.env`.
3. Set `KERBEROS_PASSWORD` to a long random secret.
4. Set `SECURE_SERVER_LOCATION` to a public host/IP that can actually be reached by test consoles.
5. Run `docker compose up --build -d`.

This starts both UDP server processes and MongoDB. The optional internal token bridge is disabled unless its secret and public auth host are configured. The stack does **not** create Pretendo Network IDs, write public DNS records, or alter hosted account-server behavior. The optional account-server patch is only for an account service you operate or are authorized to modify.

## Provisioning a NEX test record

The account service uses `NEX_ACCOUNTS_DATABASE` and `NEX_ACCOUNTS_COLLECTION`. By default, the local deployment uses the independent `wsc.nexaccounts` collection. It does not automatically contain Pretendo accounts.

For an **existing PID** in a test environment, set `WSC_NEX_PASSWORD` in your shell to the NEX credential expected for that test account and run:

```sh
export WSC_NEX_PASSWORD='your-test-nex-password'
docker compose run --rm --entrypoint ./account-admin -e WSC_NEX_PASSWORD wsc-authentication -pid 123456
unset WSC_NEX_PASSWORD
```

Replace `123456` with the actual existing 32-bit PID. The tool upserts that PID and password into the configured collection and does not print the password. This does **not** create a PNID, assign a PID to a Wii U, or connect this project to Pretendo's account service. Do not use production credentials for local tests.

Only set `NEX_ACCOUNTS_DATABASE=pretendo` and `NEX_ACCOUNTS_COLLECTION=wsc_nexaccounts` if you administer or have explicit authorized access to a compatible account database.

## Development status

- [x] Add attributed authentication and secure-server code.
- [x] Make NEX account database/collection configurable and fail clearly if the DB is unavailable.
- [x] Add a test-only account provisioning utility for existing PIDs.
- [x] Add resolver-specific DNS preflight tooling and unit tests.
- [x] Record sourced WSC NEX ID/access key, account-server token route, response fields, known title IDs, and historical Nintendo endpoint.
- [x] Add endpoint metadata checks to CI.
- [x] Add Compose validation and container builds to CI.
- [x] Add an optional, WSC-only account-route integration patch for a self-hosted/authorized Pretendo Account Server.
- [x] Resolver lookups through Pretendo and WiiLink DNS completed successfully (maintainer-reported); this does not confirm account-route behavior.
- [ ] Deploy the route patch on an account server we control or obtain provider-side routing for game-server ID `1012F100`.
- [ ] Capture Wii U startup traffic and confirm the account response advertises this project's authentication host/port.
- [ ] Verify login and Wii Sports Club traffic through each DNS profile independently.
- [ ] Add captured-packet regression tests for login, club data, rankings, matchmaking, and each sport.
- [ ] Test on Wii U hardware and document which regions and modes are verified.

## References

- [Pretendo Wii U setup](https://pretendo.network/docs/install/wiiu)
- [WiiLink DNS guide](https://wiilink.ca/guide/dns/)
- [Pretendo Network NEX overview](https://nintendo-wiki.pretendo.network/docs/nex/)
- [Endpoint reference maintained in this repo](docs/ENDPOINT-REFERENCE.md)
- [Wii U account-server token API documentation](https://github.com/kinnay/NintendoClients/wiki/Account-Server)
- [Upstream Wii Sports Club server source](https://github.com/BrewtendoNetwork/Wii-Sports-Club)
