# Online readiness and remaining hosting steps

This guide records exactly what is ready and what still must be configured before an unmodified Wii U can use this server. It assumes the single-container Docker Desktop image has already been published to your public Docker Hub repository.

## Current state

- The authentication and secure/game servers build in CI.
- The single-container desktop image packages MongoDB and both WSC NEX servers; the current CI has successfully run its startup/health smoke test.
- The desktop image can be started through Docker Desktop's graphical **Run** dialog without Terminal.
- A WSC-only route patch exists for a **self-hosted account server you operate or are explicitly authorized to modify**. It uses the authenticated account's resolved PNID/PID and calls the private WSC token bridge.
- The route patch is applied and TypeScript-checked by the account integration CI against the documented Pretendo Account Server baseline.

**Not yet verified:** a live Wii U login, a public NEX session, or a real online match. No public account/service-discovery route is deployed by this repository.

## Why running the game server is not sufficient

Wii U does not discover its NEX authentication server just by looking up a game-server DNS record. It asks the account service for the NEX token at:

`GET /v1/api/provider/nex_token/@me?game_server_id=1012F100`

That account response must return the community authentication hostname on UDP 60014, plus the authenticated PID and NEX credentials. After that handshake, this project's authentication server advertises the secure/game server on UDP 60015.

The route patch is at [provider-route.patch](../integrations/pretendo-account/provider-route.patch). It cannot be installed on hosted Pretendo Network or WiiLink services from this repository, and this repository cannot edit their public DNS or account-server routes. **Using Pretendo DNS or WiiLink DNS by itself does not redirect the WSC account-service response to this project.**

There are only two legitimate ways to bridge this final gap:

1. The operator of the account service used by the console deploys/authorizes the WSC route to this server.
2. You operate a compatible account-service stack yourself and configure your own client/service-discovery environment. That is a separate task from running the NEX game servers: it requires account/PID handling, HTTPS/client-certificate handling or an authorized client patch, and the right DNS/service routing.

Do not redirect unrelated Nintendo domains, collect other people's credentials, or expose the private bridge to the Internet.

## A. Start your published desktop image on the Mac

These steps only start the local backend; they do not finish service discovery.

1. Open Docker Desktop and make sure the engine is running.
2. Open **Images**, find the image you published at `YOUR_DOCKERHUB_USERNAME/pretendo-wii-sports-club-desktop`, and pull its `main` tag.
3. Click **Run** and expand **Optional settings**. Use container name `wsc-local`.
4. Map **UDP 60014 → host UDP 60014** and **UDP 60015 → host UDP 60015**.
5. Keep the image's `/data` persistent volume. Do not publish TCP 8080 or MongoDB port 27017.
6. Set `KERBEROS_PASSWORD` to a long unique test secret.
7. Set `SECURE_SERVER_LOCATION` to `localhost` for process-start testing. If later testing with a Wii U on the **same permitted LAN**, change it to the Mac's LAN IPv4 address (view it through **System Settings → Network → Wi-Fi → Details → TCP/IP**). Do not use `localhost` as an address for the Wii U; that would mean the console itself.
8. Click **Run**, then open **Containers → wsc-local → Logs**. Confirm MongoDB, the secure server, and the auth server start without errors.

The school Mac and school network may be managed and may prohibit running network servers or accepting inbound UDP. Do not bypass those controls. A healthy Docker container is useful local progress, but it is not evidence that a remote Wii U can connect.

## B. Hosting on a reachable network

For another player to connect, the server must remain running and the address advertised by the service must be reachable from their console.

- For a **local-LAN test**, the Mac and Wii U must be on the same permitted network, UDP 60014/60015 must be allowed, and the route must advertise the Mac's LAN address.
- For **Internet hosting**, a continuously available host needs a public IP or DNS name and provider/router firewall rules allowing inbound UDP 60014 and UDP 60015. Configure `SECURE_SERVER_LOCATION` as the publicly reachable secure-server address and configure `NEX_AUTH_PUBLIC_HOST` in the auth/bridge environment as the publicly reachable authentication address. They may be the same IP or DNS name.
- A school Mac is usually not a suitable public host: campus networks often block inbound traffic, and sleep/reboots would stop the service. Do not attempt router/firewall changes on a managed network without authorization. With no hosting budget, the alternatives are an eligible genuinely-free continuously running host (verify its network rules first) or a volunteer community host.

## C. Configure the WSC route on an account service you control

For a self-hosted/authorized Pretendo Account Server based on the baseline referenced by this repo:

1. Apply [the WSC route patch](../integrations/pretendo-account/provider-route.patch) to `src/services/nnas/routes/provider.ts`. The [account integration CI](../.github/workflows/account-integration-patch.yml) verifies application and TypeScript compilation for its pinned upstream baseline.
2. Ensure the account-service process and WSC auth process can reach each other over a **private container network**.
3. Configure `WSC_INTERNAL_API_URL=http://wsc-authentication:8080/internal/wsc/nex-token` if using that shared Docker network. Otherwise set it to an equivalent private-only URL that is reachable from that account service.
4. Set a long random `WSC_INTERNAL_API_TOKEN` (at least 32 characters) in both the WSC auth process and account server. Never put this token in the browser-visible image or publish TCP 8080 externally.
5. Set `NEX_AUTH_PUBLIC_HOST` to the real hostname/IP reachable from the Wii U and `NEX_AUTH_PUBLIC_PORT=60014`. Set `SECURE_SERVER_LOCATION` to the real secure-server hostname/IP and `SECURE_SERVER_PORT=60015`.
6. Use a dedicated test PNID and verify the account route returns XML with the expected auth host/port and the correct account PID. Redact all tokens/passwords before saving or sharing logs.

This route patch is **not** a complete standalone Nintendo account server. If you choose to operate your own account server rather than obtain a provider-side route, you must separately finish its deployment, user/account setup, compatible HTTPS/client-certificate handling, DNS/service-discovery routing, and any client-side configuration needed by your console. Those components are not supplied or publicly deployed by this repo.

## D. Prove the flow one step at a time

Only after the server address and account route have been configured:

1. Check that the account response advertises your NEX auth host/UDP 60014.
2. Check that the auth server replies with the correct secure station host/UDP 60015.
3. Test a dedicated Wii U account login from a permitted network.
4. Confirm club/profile data, then rankings, then matchmaking, and each sport separately.
5. Repeat with each intended DNS profile only if the service-discovery route for that profile actually targets your account service.

Keep a record of the game region/title ID and outcome, but redact PNIDs, passwords, tokens, tickets, certificates, and raw authenticated traffic.

## Current blocker summary

The code-side bridge and route patch are prepared. The blocking dependency is **who serves the Wii U account/service-discovery endpoint and can configure this WSC route**. Until that endpoint is under your control or the provider explicitly enables the route, there is no honest way to mark public online play as working.
