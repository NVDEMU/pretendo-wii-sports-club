# Wii U service discovery using existing Pretendo accounts

## Existing account flow

Wii Sports Club requests a NEX token for game-server ID `1012F100` through the account service. The player's existing Pretendo PNID is authenticated first; the WSC route resolves the associated NEX account and uses that account's NEX PID. No player account is created, copied, migrated, or changed.

For this WSC-only route, the account service returns the community NEX authentication host on UDP 60014, a WSC-specific stable password derived from the NEX PID, and a short-lived registered NEX token. The WSC authentication server derives the same password locally. It then advertises the secure/game server on UDP 60015.

References: [Pretendo NEX overview](https://nintendo-wiki.pretendo.network/docs/nex/), [NintendoClients Account Server documentation](https://github.com/kinnay/NintendoClients/wiki/Account-Server), [WSC NEX metadata](https://github.com/PretendoNetwork/nex-viewer/blob/master/src/nex/titles/wii-sports-club.ts).

## One WSC-only route registration

The proposed [provider-route.patch](../integrations/pretendo-account/provider-route.patch) handles only game-server ID `1012F100`; the existing generic handler remains in place for all other games. It registers the NEX token and returns the configured WSC auth host/port. This is the service-side route/registration change needed for a new fan server; it does not modify individual user accounts.

The account route and WSC auth service both use `WSC_NEX_PASSWORD_SECRET`. It must be the same random secret of at least 32 bytes in both places. Both sides compute HMAC-SHA256 with the secret's UTF-8 bytes as the key and the NEX PID's base-10 decimal string as the message. The account service uses `WSC_NEX_AUTH_HOST` and `WSC_NEX_AUTH_PORT=60014`; the WSC auth server uses `SECURE_SERVER_LOCATION` and `SECURE_SERVER_PORT=60015` for the secure station URL.

No custom domain, HTTPS callback, or separate token bridge is required for this design. The secret is server configuration, not a player's PNID password. Never commit or publicly share it.

## Reachability

Public online play still requires a continuously running host and inbound UDP 60014/60015. A LAN IP works only when the Wii U is on a permitted network that can reach it. The school Mac may be used for local tests if the network allows it, but GitHub Actions is not a permanent UDP host.

## Validation boundary

CI can validate the code and patch. It cannot deploy changes to Pretendo's live account service. The Pretendo operator (or a server operator acting with authorization) must approve/deploy the route or register an equivalent endpoint for game-server ID `1012F100`. A real Wii U login and online match remain unverified until this registration is active and a console test succeeds.
