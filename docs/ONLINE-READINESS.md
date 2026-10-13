# Online readiness with existing Pretendo accounts

The intended design uses each player's **existing Pretendo PNID**. There is no need to create or migrate accounts.

## Prepared in this repository

- WSC-specific account route patch for game-server ID `1012F100`.
- Token registration in the Pretendo account service.
- Stable WSC-specific NEX password derived from the existing NEX PID with HMAC-SHA256.
- Matching password derivation in the WSC authentication server.
- Compose/Docker environment wiring for `WSC_NEX_PASSWORD_SECRET`.

## What still must happen

1. A host must run continuously at a host/IP reachable by the Wii U, exposing UDP 60014 and 60015.
2. The account-service operator must register/deploy the WSC-specific route (or another supported endpoint registration) and set `WSC_NEX_AUTH_HOST`, `WSC_NEX_AUTH_PORT=60014`, and the shared `WSC_NEX_PASSWORD_SECRET`.
3. The WSC auth container must use the exact same shared secret and advertise the reachable secure/game host with `SECURE_SERVER_LOCATION` / UDP 60015.
4. Test login and an actual match on a Wii U.

This is one **service-side route/registration**, not a change to players' PNIDs or account records. The account service needs to tell WSC where the new server is; using existing accounts alone cannot make it discover a new endpoint.

## No domain requirement

The NEX response uses host/IP and port fields. No DuckDNS domain or public HTTPS token bridge is required by the shared-HMAC design. The school Mac can be used for process startup and possibly same-LAN testing if the network permits inbound traffic. It is not a public host unless its network makes it reachable from the Internet.

See [service discovery](SERVICE-DISCOVERY.md), [hosting instructions](OCI-WSC-HOSTING.md), and [Docker Desktop GUI startup](DOCKER-DESKTOP-GUI.md).

## Verification status

CI can test the HMAC calculation and account-route patch application. It cannot deploy the patch to Pretendo's hosted service. A live console login, accepted token, matchmaking session, and online match remain unverified until the route is deployed and the Wii U is tested.
