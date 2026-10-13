# Self-hosted account-server integration

The Wii U obtains its NEX authentication endpoint from the account server route:

`GET /v1/api/provider/nex_token/@me?game_server_id=1012F100`

A successful DNS lookup only proves that the hostname resolves; it does not prove that this route returns this project's endpoint.

## Route behavior and title IDs

The route patch checks the game-server ID `1012F100` separately from the Wii U application's title ID. It accepts the candidate WSC title IDs recorded in [the endpoint metadata](../../config/wii-sports-club-endpoints.json), requires the authenticated PNID to have a Pretendo NEX account, and passes that account's `nexAccount.pid` to the WSC bridge. Pretendo's PNID PID (`pnid.pid`) and game/NEX PID (`nexAccount.pid`) are distinct fields and must not be confused. Client-supplied PIDs are never accepted. The generic route for other game-server IDs stays intact. The allowlist contains candidate regional/version IDs because public title maps differ; capture the actual `x-nintendo-title-id` from your own test request, redact credentials, and verify that ID before broadening the allowlist.

The [Account-server integration patch check](../../.github/workflows/account-integration-patch.yml) applies this patch to a pinned upstream baseline and runs the account server's TypeScript compiler. That validates patch application and typing, but it does not deploy the route or prove a live Wii U request.

Pretendo Account Server has a route in `src/services/nnas/routes/provider.ts`. In that implementation, the account server selects a registered server based on `game_server_id`, then returns its host/port plus NEX account credentials. For a self-hosted or explicitly authorized account server, the next integration task is to route only WSC game server ID `1012F100` to this project's private token bridge, leaving all other games on the existing route.

## Required configuration for an authorized deployment

- Configure `WSC_INTERNAL_API_TOKEN` with a long random secret in this project's environment and the same secret in the account-server process.
- Configure `NEX_AUTH_PUBLIC_HOST` and `NEX_AUTH_PUBLIC_PORT=60014` for the public NEX auth endpoint.
- Set `SECURE_SERVER_LOCATION` for the public secure-server endpoint and expose UDP 60014 and 60015.
- Make the account-server process reach `http://wsc-authentication:8080/internal/wsc/nex-token` over a private network only. If they are not on the same Docker network, use a private address reachable only by the account server. Never publish the helper port to the public Internet.
- Test using a dedicated test account and verify the returned XML and NEX login before trying actual matches.

The private bridge provisions a random NEX account password on first use of a NEX PID and reuses it for later requests. A unique MongoDB index makes each PID map to one password even during concurrent first logins, and a retry handles the expected duplicate-key race. The bridge returns a separate random XML `token` value for each request. This prevents concurrent token requests from changing the password during login. Treat the bridge as an internal service; it is not the Wii U-facing account API.

## Public DNS-provider limitation

This repository cannot change Pretendo's or WiiLink's hosted service-discovery/account-server behavior or public DNS records. To support users of public Pretendo DNS or public WiiLink DNS, the relevant provider-operated account-server path has to route WSC to a compatible endpoint. This guide only applies to a server operator who can make that change; DNS lookups that succeed do not prove the route is configured.
