# Hosting WSC with existing Pretendo accounts

**Players keep their existing Pretendo PNIDs.** This setup does not create a replacement account system or change individual account records.

## What is needed

- A machine that stays running and has a host/IP reachable from the Wii U.
- Inbound UDP ports 60014 (NEX authentication) and 60015 (secure/game service).
- A WSC-only account-service route or supported server-registration entry for game-server ID `1012F100`.
- One shared server-side `WSC_NEX_PASSWORD_SECRET`, configured on both the WSC auth container and the account route. This is not a player's password.

No custom DNS name, DuckDNS token, HTTPS token bridge, or second account system is needed by this design.

## Host choices

### School Mac

Docker Desktop can run the stack for startup tests. If the Wii U is on the same permitted LAN, map UDP 60014/60015 and set `SECURE_SERVER_LOCATION` to the Mac's LAN IPv4 address. Do not use `localhost` for the console. School policy/network may block inbound traffic; do not bypass those restrictions. The Mac will not be a dependable public host if it sleeps, leaves the network, or loses connectivity.

### Public host

A public IP address is sufficient for the NEX endpoint; a domain is not required for these UDP services. Since there is no hosting budget, use a genuinely free eligible host only if it already shows no charge, or ask a community member to host. Do not buy a VPS or enter billing information just for this project.

Open UDP 60014 and UDP 60015 in both the provider firewall and the machine firewall. Do not expose MongoDB 27017 or optional internal TCP 8080.

## Configure the WSC host

In the WSC host's `.env`:

```dotenv
KERBEROS_PASSWORD=<long random server/Kerberos secret>
SECURE_SERVER_LOCATION=<public IP/hostname the Wii U can reach>
SECURE_SERVER_PORT=60015
WSC_NEX_PASSWORD_SECRET=<shared secret configured for the WSC account route>
```

The same random `WSC_NEX_PASSWORD_SECRET` must be used in the account route and WSC auth container. Use at least 32 bytes; a 64-character hex value is recommended. Do not commit it to Git or post it in an issue.

On the authorized account service, set:

```dotenv
WSC_NEX_AUTH_HOST=<public IP/hostname of the WSC authentication server>
WSC_NEX_AUTH_PORT=60014
WSC_NEX_PASSWORD_SECRET=<the exact same shared secret>
```

These account-service values belong on the service or authorized deployment, not in public source. If the account route is not deployed/registered, existing Pretendo PNIDs won't be directed to the custom server.

## Bootstrap

The optional [bootstrap script](../deploy/oci-wsc-bootstrap.sh) installs Docker and starts the WSC containers. It requires the already-configured shared secret and a host with a reachable IP. It does not require a domain or HTTPS bridge.

## Verification

1. Confirm both containers remain healthy.
2. Confirm UDP 60014 and UDP 60015 are open and externally reachable.
3. Confirm the account response for `game_server_id=1012F100` returns the existing NEX PID, WSC auth host/port, derived WSC password, and registered token.
4. Confirm the auth response advertises the secure/game host on UDP 60015.
5. Test a real Wii U login with an existing Pretendo PNID, then test club data, rankings, matchmaking, and a match.

CI/build success is not proof of live online play. Do not claim the game works online until a real console test passes.
