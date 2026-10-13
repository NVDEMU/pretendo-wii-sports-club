# Pretendo account-route integration for existing PNIDs

This patch is intended for the existing Pretendo account service. It does **not** create a replacement account network and does not change individual players' PNID/NEX account records.

The Wii U requests a NEX token from:

`GET /v1/api/provider/nex_token/@me?game_server_id=1012F100`

For this WSC-only game-server ID, the route uses the already-authenticated PNID and resolves its existing NEX account/PID. It verifies the WSC title ID, derives a WSC-specific password, creates and registers a short-lived NEX token, then returns the community auth host/port in the XML response. All other game routes remain unchanged.

## Required account-service configuration

On the authorized account-service deployment, configure:

- `WSC_NEX_AUTH_HOST`: public host/IP of the WSC authentication server.
- `WSC_NEX_AUTH_PORT=60014`.
- `WSC_NEX_PASSWORD_SECRET`: random value of at least 32 bytes (a 64-character hex value is recommended).

The **exact same** `WSC_NEX_PASSWORD_SECRET` must be configured in the WSC authentication container. Both sides use the secret's UTF-8 bytes as an HMAC-SHA256 key and the NEX PID in base-10 decimal as the message. Different values cause login failure because the client and server derive different passwords.

The secret is a server-to-server configuration value, not the player's PNID password. Never commit it, paste it into a public issue, or include it in screenshots.

## WSC host configuration

The WSC host uses:

```dotenv
WSC_NEX_PASSWORD_SECRET=<same shared secret as the account service>
SECURE_SERVER_LOCATION=<reachable host/IP of the secure/game server>
SECURE_SERVER_PORT=60015
```

The account service's route advertises its configured auth host/IP on UDP 60014; the auth server then advertises the secure/game server on UDP 60015.

No callback domain, DuckDNS name, or HTTPS token bridge is required by this design. A host that stays running and allows inbound UDP 60014/60015 is still necessary for Internet play.

## Operator boundary and validation

The [provider-route.patch](provider-route.patch) is a proposed WSC-only service integration. It can be applied by the Pretendo account-service operator or in an account-service deployment you are authorized to modify. This repository cannot push configuration into Pretendo's live service; the operator must approve/deploy the route or register an equivalent endpoint.

The [account integration workflow](../../.github/workflows/account-integration-patch.yml) checks patch application and TypeScript compilation against the pinned upstream baseline. Passing those checks does not mean the hosted route has been deployed or that a Wii U login has succeeded. After deployment, verify the XML response using a test account, then test NEX authentication, club/profile data, rankings, matchmaking, and an actual match.

Do not share players' PNID passwords, NEX passwords, tickets, tokens, or certificates in logs or issues.
