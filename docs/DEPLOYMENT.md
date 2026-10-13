# Deployment and online verification

This repo runs the Wii Sports Club NEX authentication and secure/game services. Players keep their existing Pretendo PNIDs.

## Network

- Authentication server: UDP 60014.
- Secure/game service: UDP 60015.
- MongoDB 27017 and optional helper TCP 8080 must remain private.
- No custom domain or HTTPS bridge is required by the shared-HMAC design.

## Configure the WSC host

In `.env`:

```dotenv
KERBEROS_PASSWORD=<long random server/Kerberos secret>
SECURE_SERVER_LOCATION=<public host/IP the Wii U can reach>
SECURE_SERVER_PORT=60015
WSC_NEX_PASSWORD_SECRET=<shared secret configured for the WSC account route>
```

The shared secret must match the authorized account-route configuration. It must be at least 32 bytes; a 64-character hex value is recommended. Do not commit or share it publicly.

The account service must set `WSC_NEX_AUTH_HOST=<public host/IP>`, `WSC_NEX_AUTH_PORT=60014`, and the exact same `WSC_NEX_PASSWORD_SECRET`.

## Deploy

Use Docker Compose on a machine that stays running. For public online play, permit inbound UDP 60014 and 60015 through the host firewall and provider firewall. A Mac on a LAN is only usable from devices that can reach it; `localhost` is for process-start tests only.

## Existing Pretendo account integration

The [WSC provider route patch](../integrations/pretendo-account/provider-route.patch) intercepts only game-server ID `1012F100`, keeps other game routes unchanged, uses the authenticated user's existing NEX PID, registers a NEX token, and returns the WSC endpoint/password. The route is a proposed integration and is not deployed to Pretendo's running service by this repository. An account-service operator must approve/deploy the route or configure an equivalent supported game-server registration. This is service configuration, not a change to individual accounts.

## Validate

1. Run CI and verify container health.
2. Confirm account response returns the existing NEX PID, WSC auth host/port, password, and registered token.
3. Confirm auth advertises the reachable secure host on UDP 60015.
4. Test a real Wii U login, club/profile data, rankings, matchmaking, and a match. CI is not a substitute for live-console validation.
