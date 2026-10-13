# Wii Sports Club Network Revival

A community project exploring replacement online services for **Wii Sports Club on Wii U**.

> **Status: work in progress.** This repository is not affiliated with Nintendo, Pretendo Network, or WiiLink. The services compile and container images build, but end-to-end Wii U login and online play have not yet been verified.

> **Budget note:** no paid hosting is required to continue development. Follow the [zero-cost plan](docs/FREE-PLAN.md) to keep working through GitHub's web editor and free CI. Public online play needs a continuously reachable server, which remains on hold until genuinely free hosting or a volunteer host is available.

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
- `POST /internal/wsc/nex-token` on private TCP 8080 — optional local/developer helper, not required by the intended public Pretendo account route.
- `docker-compose.yml` — local deployment scaffold with MongoDB.
- `docs/MACOS-QUICKSTART.md` — **Mac-first local setup and browser-based Codespaces setup; no local Linux installation required.**
- `scripts/macos-local-start.sh` — creates a local-only environment, starts Compose, and runs the internal-bridge smoke test.
- `docs/DOCKER-DESKTOP-GUI.md` — run a single-container local bundle through Docker Desktop's graphical Run dialog, with no Terminal.
- `docs/ONLINE-READINESS.md` — remaining account/service-discovery requirements and post-publish hosting steps.
- `docs/OCI-WSC-HOSTING.md` — optional host guide for existing Pretendo PNIDs; no custom domain or HTTPS token bridge required.
- `docs/SERVICE-DISCOVERY.md` — WSC-only account-route requirement; individual Pretendo accounts remain unchanged.
- `deploy/oci-wsc-bootstrap.sh` — browser/cloud-init bootstrap for the WSC host.
- `deploy/Dockerfile.desktop` — bundles MongoDB and both WSC servers in a single image for GUI-only local testing.
- `deploy/start-wsc-desktop.sh` — starts and supervises the local bundle processes.
- `config/wii-sports-club-endpoints.json` — machine-readable endpoint metadata with source URLs and confidence notes.
- `docs/ENDPOINT-REFERENCE.md` — sourced game-server ID, access key, account API route, historical Nintendo endpoint, and community ports.
- `docs/DEPLOYMENT.md` — public-server hosting, firewall, private account-service integration, and hardware verification sequence.
- `docs/BROWSER-DEPLOYMENT.md` — browser-only cloud deployment guide; no local Terminal required.
- `docs/FREE-PLAN.md` — zero-cost development/test roadmap; use this before considering any cloud host.
- `deploy/digitalocean-cloud-init.yaml` — first-boot setup to paste into a Droplet's Startup scripts field.
- `.github/workflows/publish-containers.yml` — publishes runtime images only after main-branch CI passes.
  Optional Docker Hub publishing is enabled by adding `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` as repository Actions secrets, so the desktop image can be pulled through Docker Desktop's GUI.
- `docker-compose.cloud.yml` — cloud runtime config with prebuilt images and persistent data volumes.
- `scripts/smoke_test_internal_bridge.py` — end-to-end check of the internal token bridge against MongoDB.
- `docs/SERVICE-DISCOVERY.md` — describes the Wii U account/service-discovery requirement.
- `integrations/pretendo-account/provider-route.patch` — WSC-only provider-route integration for an authorized Pretendo account-service deployment; it uses existing PNIDs and a shared HMAC secret. See the [integration guide](integrations/pretendo-account/README.md).
- `scripts/check_dns_routing.py` — queries a selected resolver for a verified hostname; can check the expected IPv4 result.
- `.github/workflows/ci.yml` — Go builds, endpoint/DNS unit tests, Compose validation, container builds, and a Compose/MongoDB bridge smoke test.
- `.github/workflows/dns-preflight.yml` — manually query a verified hostname through both public DNS resolvers.

The server components are based on the GPL-3.0-licensed Wii Sports Club server code published by [BrewtendoNetwork/Wii-Sports-Club](https://github.com/BrewtendoNetwork/Wii-Sports-Club). The upstream license and attribution are retained. This is a separate community-maintained project, not an official Pretendo release.

## Easiest local test (Docker Desktop)

If Docker Desktop is already installed, start with the [Docker Desktop GUI guide](docs/DOCKER-DESKTOP-GUI.md). It walks through running the bundled authentication server, game server, and database without installing Go or configuring multiple containers. For now, use this as a **local process/startup test**; it is not yet a one-click online connection for a Wii U.

**Player experience goal:** connect the same way as any other Pretendo-supported game: use the player's existing Pretendo account, no separate revival registration, no second password prompt, and no manual per-player account setup. The revival should handle WSC-specific credentials behind the scenes. The remaining blocker is service discovery: the Wii U must be told to connect to this revival's server for WSC. A local Docker test does not change Pretendo's hosted account route or public DNS, and it does not prove real console login yet.

## Browser-only quick start (no Terminal required)

You can build and test repository changes entirely from GitHub's website. Open the [Browser-only quick start](docs/BROWSER-ONLY-QUICKSTART.md), then visit [GitHub Actions](https://github.com/NVDEMU/pretendo-wii-sports-club/actions). Select **WSC Build & Integration Tests (browser-run)**, click **Run workflow**, keep branch `main), and click **Run workflow** again. The tests also run automatically after every commit.

This hosted workflow compiles and tests both services, checks DNS/configuration logic, builds the containers, and smoke-tests the internal NEX bridge against a temporary MongoDB container. No local Terminal, Go, Python, or Linux installation is needed.

For developers who do have Terminal available, see the [macOS/Codespaces quickstart](docs/MACOS-QUICKSTART.md).

**Important:** GitHub's test environment is temporary. Passing tests does not deploy a public server or prove Wii U login/online play. Actual online tests still need a public host with inbound UDP 60014/60015 and an authorized, compatible account/service-discovery route.

## General Docker Compose setup

For local development, Docker Compose starts MongoDB and both NEX services. For a public server, configure `KERBEROS_PASSWORD`, `SECURE_SERVER_LOCATION`, and `WSC_NEX_PASSWORD_SECRET` in `.env` before deploying the containers. The shared secret must match the authorized Pretendo account-route configuration exactly.

**Existing Pretendo accounts are the intended player sign-in path.** The WSC-only provider route resolves the already-authenticated PNID's NEX PID, registers a short-lived token, and derives a WSC-specific password with HMAC-SHA256. The WSC authentication server derives the same password from `WSC_NEX_PASSWORD_SECRET`. No player PNID or individual account record is created, copied, or changed.

The remaining provider-side action is to approve/deploy the route (or register an equivalent server endpoint) for `game_server_id=1012F100`. This is a service-discovery registration, not a player account modification. This repository cannot change Pretendo's hosted account service, and configuring Pretendo DNS alone does not make the console receive this project's endpoint. No custom domain or HTTPS callback tunnel is required by the HMAC design.

The private TCP 8080 NEX-token bridge remains as a local/developer helper for tests; it is not part of the intended public Pretendo account route. Keep it un-published.
 
## Provisioning a NEX test record

The account service uses `NEX_ACCOUNTS_DATABASE` and `NEX_ACCOUNTS_COLLECTION`. By default, the local deployment uses the independent `wsc.nexaccounts` collection. It does not automatically contain Pretendo accounts.

**This manual command is only for an isolated/local NEX test.** It is not the intended sign-in flow for players using Pretendo IDs. In the intended integration, the account service authenticates the existing PNID and resolves its NEX/game PID; the route and auth server derive the same WSC-only NEX password from a shared secret. Do not ask players for their Pretendo password, and do not point this revival at Pretendo's private database.

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
- [x] Add a real MongoDB/Compose internal-bridge smoke test to CI.
- [x] Add a one-command local smoke-test path for macOS and a Codespaces devcontainer.
- [x] Add a WSC-only account-route patch that keeps existing Pretendo PNIDs unchanged and uses shared HMAC-derived game credentials.
- [x] Resolver lookups through Pretendo and WiiLink DNS completed successfully (maintainer-reported); this does not confirm account-route behavior.
- [ ] Deploy the route patch on an account server we control or obtain provider-side routing for game-server ID `1012F100`.
- [ ] Obtain a public UDP-capable server endpoint and verify the UDP ports from an external network.
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
