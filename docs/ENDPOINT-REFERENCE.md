# Wii Sports Club endpoint reference

This file records endpoint values recovered from public documentation and server implementations so we can distinguish the game identifier, the original Nintendo endpoint, and the endpoints for this community deployment.

Machine-readable copy: [`config/wii-sports-club-endpoints.json`](../config/wii-sports-club-endpoints.json).

## Confirmed game/server identity

| Field | Value | Meaning |
| --- | --- | --- |
| Wii Sports Club NEX game-server ID | `1012F100` | Value sent as the `game_server_id` query parameter in the account-server token request |
| NEX access key | `4d324052` | Public game-specific NEX access-key value found in the game metadata and existing WSC server implementation |
| PRUDP protocol | V1; configured minor version 3 | NEX transport setting used by the community server code |
| Community server NEX version | `3.4.0` | Version configured by the current authentication and secure-server code |
| Community NEX authentication port | UDP `60014` | Custom deployment port; must be returned by service discovery |
| Community NEX secure port | UDP `60015` | Custom deployment port advertised by the authentication server's secure station URL |

Sources: [Pretendo's Wii Sports Club NEX metadata](https://github.com/PretendoNetwork/nex-viewer/blob/master/src/nex/titles/wii-sports-club.ts) and [the public Revivetendo WSC configuration](https://github.com/Happynico7504/RevivetendoOpenSource/blob/master/relaylink/nexgames.go). The ports 60014/60015 are community deployment choices, not Nintendo's original port.

## Wii U service-discovery request

The documented Wii U flow asks the account server for its NEX token at:

```http
GET https://account.nintendo.net/v1/api/provider/nex_token/@me?game_server_id=1012F100
```

The Wii U account server documentation describes the common client certificate required for account-server requests. [Source: NintendoClients Account Server wiki](https://github.com/kinnay/NintendoClients/wiki/Account-Server).

The documented XML response fields are:

```xml
<nex_token>
  <host>community-auth.example.net</host>
  <port>60014</port>
  <pid>EXISTING_PID</pid>
  <nex_password>SESSION_NEX_PASSWORD</nex_password>
  <token>SESSION_TOKEN</token>
</nex_token>
```

This snippet is a **schema example**, not a usable hostname, PID, credential, or live response. The response needs a reachable community authentication endpoint and a valid account/PID credential arrangement. An account proxy/compatible service-discovery integration must securely map the authenticated player's account to the intended NEX PID and credential, and then return the community authentication host/port. The current repository does **not** implement the upstream account server or host this route.

Once the Wii U reaches the NEX authentication server, this repo's authentication code advertises the secure server using `SECURE_SERVER_LOCATION` and `SECURE_SERVER_PORT`. These are separate endpoints in the protocol flow:

1. Account/service discovery returns the **NEX authentication server** on UDP 60014.
2. The NEX authentication server replies with a secure station URL pointing to the **NEX secure server** on UDP 60015.
3. The client then uses the secure server for game services.

See [SERVICE-DISCOVERY.md](SERVICE-DISCOVERY.md) for the missing integration boundary.

## Historical Nintendo endpoint — do not use for the community server

The reverse-engineered Wii U NEX title list in [`kinnay.github.io/data/nexwiiu.json`](https://github.com/kinnay/kinnay.github.io/blob/master/data/nexwiiu.json) records:

| Field | Archived value |
| --- | --- |
| Server address | `34.208.166.202` |
| UDP port | `43480` |
| Game-server ID | `1012F100` (decimal 269676800) |
| Access key | `4d324052` |
| Branch | `branch:origin/project/appsp` |
| Client build | `build:3_4_24_4_0` |

This is historical Nintendo endpoint data. It is **not** the destination for the community server, and the original Nintendo service must not be expected to work: Nintendo discontinued online play and other online-communication features for Wii U software on April 8, 2024 ([official notice](https://en-americas-support.nintendo.com/app/answers/detail/a_id/63227/p/605)).

The client build string `3_4_24_4_0` is not the same field as the server library's NEX protocol version `3.4.0`; do not mix them.

## Title IDs and regional variants

The Pretendo NEX Viewer lists these title IDs: `000500001012F100`, `0005000010144D00`, `0005000010144E00`, and `0005000010171E00`. A separate public Revivetendo title map also includes `0005000010144C00`. The source sets differ, so the complete region/version mapping remains a **verification item**. Do not infer a region from the suffix alone; confirm the installed title ID from the console/game metadata or a redacted request capture.

## DNS is not service discovery

Current documented Wii U resolver addresses:

- Pretendo SSSL: `88.198.140.154`
- WiiLink DNS: `167.235.229.36`

Those addresses are DNS resolvers/entry points, not our game server. A resolver can route a hostname, but it cannot by itself change the XML response from the account-server API. Both DNS use cases require a service-discovery path that actually returns this project's NEX authentication address. This repository cannot change Pretendo's or WiiLink's public records.

References: [Pretendo Wii U setup](https://pretendo.network/docs/install/wiiu), [WiiLink DNS guide](https://wiilink.ca/guide/dns/), [Pretendo NEX overview](https://nintendo-wiki.pretendo.network/docs/nex/).
