# Wii U service discovery: remaining integration boundary

## Confirmed request and game identifier

The Wii U obtains its NEX game-server location from the account-server token route, not from a Switch-style game-server DNS lookup. The documented request is:

```http
GET https://account.nintendo.net/v1/api/provider/nex_token/@me?game_server_id=1012F100
```

- Account host: `account.nintendo.net`
- Path: `/v1/api/provider/nex_token/@me`
- Wii Sports Club game-server ID: `1012F100`
- NEX access key used by the WSC server: `4d324052`

References: [Pretendo NEX overview](https://nintendo-wiki.pretendo.network/docs/nex/), [NintendoClients Account Server documentation](https://github.com/kinnay/NintendoClients/wiki/Account-Server), [WSC NEX metadata](https://github.com/PretendoNetwork/nex-viewer/blob/master/src/nex/titles/wii-sports-club.ts).

The response is XML with `host`, `port`, `pid`, `nex_password`, and `token` fields. The `host`/`port` returned here must reach the **authentication server**, not the secure server. The current community port choices are UDP 60014 for authentication and UDP 60015 for secure NEX. After authentication, this repo's server advertises the secure station URL from `SECURE_SERVER_LOCATION` and `SECURE_SERVER_PORT`.

## Private token bridge now included

The authentication container now has an optional **internal-only** route that a compatible account proxy can call after it has authenticated the Wii U account and resolved the user's existing PID:

```http
POST http://wsc-authentication:8080/internal/wsc/nex-token
Authorization: Bearer <WSC_INTERNAL_API_TOKEN>
Content-Type: application/json

{"pid":123456789}
```

It returns the XML `nex_token` response with the configured public authentication host and UDP port. It also creates a fresh random NEX session password, stores it for that PID in the configured NEX account collection, and returns it in the XML response. The same token is used for `token` and `nex_password` to match the public WSC server implementation pattern.

Enable this route only when integrating with a trusted account proxy:

- Set `WSC_INTERNAL_API_TOKEN` to a random secret of at least 32 characters (for example, generate 32 random bytes with `openssl rand -hex 32`).
- Set `NEX_AUTH_PUBLIC_HOST` to a public hostname/IP that routes to UDP 60014.
- Optionally set `NEX_AUTH_PUBLIC_PORT` (defaults to 60014).
- Keep TCP 8080 private on the Compose network. It is deliberately **not** published to the host.
- The caller must authenticate the account and provide the PID resolved for that account. Never expose this internal route directly to the internet or accept a client-supplied PID without authentication.

This helper is **not** the Wii U-facing `account.nintendo.net` service and does not modify either provider's DNS. A compatible account proxy still needs to intercept the account-server request, authenticate it, call this helper, and return the XML response to the Wii U.

## What the repo currently implements

- NEX authentication service: UDP 60014, configured for WSC access key `4d324052` and NEX version 3.4.0.
- NEX secure service: UDP 60015, configured for the same game access key and NEX version.
- The authentication server advertises a secure station URL via environment variables.
- Account PID/password lookup is backed by the configurable MongoDB account collection.

## What is still missing

This repo does **not** currently host `account.nintendo.net`, implement the account-server TLS/client-certificate flow, create or link PNIDs, or control Pretendo/WiiLink's public DNS records. As a result, an unmodified Wii U still needs a compatible account/service-discovery integration that returns this project's auth endpoint for `game_server_id=1012F100`.

DNS configuration alone is insufficient: the DNS resolver must route the HTTPS account hostname and the account response must advertise this server. The public DNS providers are operated separately; this repository cannot add routes to their deployed configuration. Avoid inventing another service hostname or copying an account endpoint from a different game.

## Implementation plan

1. Use the sourced values in [ENDPOINT-REFERENCE.md](ENDPOINT-REFERENCE.md) and [the machine-readable endpoint file](../config/wii-sports-club-endpoints.json) as the implementation baseline.
2. For a Wii U that you own, capture only the DNS names and connection metadata during a Wii Sports Club startup. The exact game request path and game ID are documented; the capture now helps confirm which resolver path the console actually takes and what happens before/after the token request.
3. Select an account/service-discovery strategy we can legitimately operate: an authorized account-server integration, or a controlled test setup that redirects/patches the service-discovery path. A plain standalone HTTP endpoint on a custom hostname is not sufficient while the unmodified client expects the Nintendo account hostname and TLS/client-certificate behavior.
4. Return a fresh per-session NEX password/token and the community authentication host/UDP port, with safe association to the authenticated user's existing PID.
5. Verify the secure station URL advertises the community secure host/UDP port.
6. Test login, club data, rankings, matchmaking, and each sport separately. Run both DNS-path tests only once the required service-discovery routing is actually available.

## Security rules for capture/testing

Use a dedicated test account. Never commit PNID passwords, device certificates, NEX passwords, tickets, bearer tokens, session tokens, or raw authenticated payloads. A packet capture can contain credentials even if the filename looks innocuous; sanitize it before sharing.
