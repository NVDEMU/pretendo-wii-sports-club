# DNS compatibility: Pretendo and WiiLink

## Goal

The NEX authentication and secure-server processes should be DNS-provider-neutral. They advertise or accept reachable server endpoints; they should not inspect the player's resolver brand or change protocol behavior based on it.

This repository aims to support players connecting through either:

- **Pretendo SSSL DNS** — Pretendo's Wii U guide currently documents `88.198.140.154`.
- **WiiLink DNS** — WiiLink's guide currently documents `167.235.229.36`. WiiLink's privacy policy says that some queries are forwarded to Pretendo Network and other services.

Official setup references:
- [Pretendo — Wii U setup](https://pretendo.network/docs/install/wiiu)
- [WiiLink — DNS guide](https://wiilink.ca/guide/dns/)
- [WiiLink — privacy policy / DNS forwarding](https://wiilink.ca/privacy-policy/)

## What this repository can and cannot control

A custom game backend does **not** automatically receive traffic just because a console uses one of these DNS resolvers. The relevant service-discovery and hostname records must point clients at the right endpoints.

This project can configure the NEX service endpoint advertised by its authentication service through `SECURE_SERVER_LOCATION` and `SECURE_SERVER_PORT`. Set the location to a public hostname or IP that reaches this deployment. Both connection paths must be able to resolve and reach it.

This repository cannot modify Pretendo's or WiiLink's hosted DNS records. If the original Wii Sports Club service hostnames themselves must be redirected, the mapping must be supplied by the selected provider or by a project-controlled DNS relay/override. We must discover and verify those exact hostnames before adding records; do not guess Nintendo service domains or redirect unrelated Nintendo services.

## Do not configure the two public resolver IPs as a pair

A Wii U connection has a primary and secondary DNS setting. The second is a fallback if the first fails; it does not merge their records or select the best response per hostname. A compatible local/hosted relay should choose one upstream profile for each run:

| Profile | Upstream |
| --- | --- |
| `pretendo` | `88.198.140.154:53` |
| `wiilink` | `167.235.229.36:53` |

Those are upstream resolver addresses, not addresses of the Wii Sports Club game server. The current repo does not include a deployed DNS relay or public overrides yet, so these profile names are documentation for deployment planning rather than a claim that direct public DNS routing is already end-to-end verified.

## Validation checklist

For each profile, verify all of the following before advertising it as working:

1. Confirm the canonical account/service-discovery hostnames from a real Wii U capture or trusted upstream documentation.
2. Confirm each relevant hostname resolves to the intended account and game endpoints when using that profile.
3. Confirm the advertised NEX secure-server host and UDP port are reachable from an external network.
4. Test authentication with a test account provisioned in the expected account store.
5. Test sport/region flows independently; Wii Sports Club uses multiple services and regional club behavior.
6. Record tested firmware, game region, DNS profile, resolver results, and packet traces in the repository without uploading passwords, tokens, or account secrets.

## Account-service caveat

The authentication service currently queries `pretendo.wsc_nexaccounts` in MongoDB for a PID/password record. The local Compose stack creates MongoDB but does not provision Pretendo IDs or bridge the Pretendo account service. Without correctly provisioned account records or an intended account-backend integration, successful game login is not guaranteed even when DNS and UDP reachability work.
