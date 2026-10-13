# Wii U service discovery: remaining integration boundary

## Why the NEX ports alone do not finish the connection

This repository includes two UDP NEX services:

- Authentication service: UDP 60014
- Secure/game service: UDP 60015

The authentication service advertises the secure server's host and port using `SECURE_SERVER_LOCATION` and `SECURE_SERVER_PORT`. Those settings control the station URL returned *after a client has reached this authentication service*.

They do not, by themselves, make an unmodified Wii U client discover or connect to this project's authentication server. Pretendo's NEX documentation notes that Wii U game-server location is requested from the account server via `/provider/nex_token/@me`. The account/service-discovery route must therefore be confirmed and configured as well.

Reference: [Pretendo Network's NEX overview](https://nintendo-wiki.pretendo.network/docs/nex/).

## Required next integration

1. Capture one clean Wii Sports Club startup with a console configured to use a known DNS profile. Record DNS queries, TCP/TLS destinations, and UDP destinations; redact account tokens and secrets.
2. Identify the service-discovery request and the exact response field that supplies the NEX station URL. Do not assume a hostname from a different game is correct.
3. Determine which routing setup the test environment uses:
   - an authorized Pretendo-compatible account/service registry that returns the community server URL;
   - a project-controlled DNS/HTTP(S) relay with the relevant host mappings; or
   - a game/client patch that selects the community endpoint.
4. Implement and test the selected path separately for Pretendo DNS and WiiLink DNS. The public resolvers are independently operated and this repository cannot change their records.
5. Test with a new or dedicated test account before using any real account. Never publish passwords, tickets, access keys, session tokens, or raw authenticated requests.

## Account store modes

The authentication service now reads its PID/password records from the configured MongoDB database and collection:

- `NEX_ACCOUNTS_DATABASE` (defaults to `wsc`)
- `NEX_ACCOUNTS_COLLECTION` (defaults to `nexaccounts`)

For a separately administered deployment, the included `cmd/account-admin` utility can add a NEX password record for an **existing PID**. It does not mint a Pretendo Network ID, change the PID associated with a console, or grant access to Pretendo's private databases.

Only configure `NEX_ACCOUNTS_DATABASE=pretendo` and `NEX_ACCOUNTS_COLLECTION=wsc_nexaccounts` when you operate or have explicit authorized access to that compatible database. A stock local Compose deployment does not contain PNID data, so it will not authenticate a PNID automatically.

## Resolver preflight

Use `scripts/check_dns_routing.py` to query the documented Pretendo or WiiLink resolver for a hostname confirmed from packet captures or trusted docs. Optionally pass `--expected-ip` to assert the returned A records contain the expected address.

The script does not invent or discover the game's private service hostnames, alter upstream DNS, or prove that a UDP NEX endpoint is reachable. A passing A-record check is one component of service-discovery validation, not an end-to-end test.
