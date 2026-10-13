# DNS compatibility: Pretendo and WiiLink

## Goal

The NEX authentication and secure-server processes are resolver-neutral. They do not inspect which public DNS brand a player uses. To work in practice, however, the selected network path must route the game's service-discovery and advertised NEX endpoints to the right services.

Official guides currently list:

- **Pretendo SSSL DNS:** `88.198.140.154` ([Pretendo Wii U setup](https://pretendo.network/docs/install/wiiu)).
- **WiiLink DNS:** `167.235.229.36` ([WiiLink DNS guide](https://wiilink.ca/guide/dns/)).

These public IP addresses are DNS resolvers, not this project's server IP. The guides establish how to configure each service's DNS, but they do not establish that our yet-unregistered Wii Sports Club hostnames are already pointed to this project. We should not claim public DNS routing is complete until tested.

## Important distinction: resolver vs service discovery

A DNS resolver translates a hostname into records. Wii U NEX service discovery also involves the account/server-location path; see [SERVICE-DISCOVERY.md](SERVICE-DISCOVERY.md). The game must actually be told to connect to the custom endpoint. Configuring `SECURE_SERVER_LOCATION` only controls the station URL advertised by this repository's NEX authentication server once the client has reached that server; it does not make the client discover the authentication service in the first place.

This repository cannot edit Pretendo's or WiiLink's public records. If those providers do not already route the relevant service hostnames to this project, a project-controlled DNS/HTTPS relay or a client/account-service integration will be required. Exact hostnames must come from trusted documentation or a packet capture; do not guess domains or redirect unrelated Nintendo services.

## Do not mix resolver IPs as a fake combined profile

The Wii U primary DNS resolver is used for lookups; the secondary resolver is generally a fallback, not a merge of both DNS databases. Setting Pretendo as primary and WiiLink as secondary does not guarantee that a hostname missing from Pretendo will be queried against WiiLink.

Test one path at a time:

| Profile | Resolver queried by the preflight tool |
| --- | --- |
| `pretendo` | `88.198.140.154:53` |
| `wiilink` | `167.235.229.36:53` |

## DNS preflight tool

Once a hostname is confirmed, query it via each profile:

```sh
python3 scripts/check_dns_routing.py --profile pretendo --hostname VERIFIED_HOSTNAME
python3 scripts/check_dns_routing.py --profile wiilink --hostname VERIFIED_HOSTNAME
```

When an expected server IPv4 address is known, assert it:

```sh
python3 scripts/check_dns_routing.py --profile pretendo --hostname VERIFIED_HOSTNAME --expected-ip 203.0.113.20
```

Replace both example values with values from the deployment and real service configuration. The tool checks an A/CNAME answer from the selected resolver; it does not update DNS, discover game hostnames, check TLS, or prove that UDP traffic reaches the NEX server.

## Validation checklist

For each DNS profile separately:

1. Capture startup traffic and confirm all account/service-discovery hostnames.
2. Verify the service-discovery response advertises this project's intended NEX host and port.
3. Query confirmed hostnames through the selected resolver and verify expected answers.
4. Check the public UDP ports from a real external network.
5. Test authentication with a dedicated account provisioned in the configured store.
6. Test sports, regional clubs, rankings, and matchmaking independently.
7. Record firmware, game region, DNS profile, resolver answers, and redacted packet notes. Never publish passwords, tickets, auth tokens, or raw authenticated payloads.
