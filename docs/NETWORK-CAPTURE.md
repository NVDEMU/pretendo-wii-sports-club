# Wii Sports Club network capture workflow

A short network capture is the next research requirement before implementing more DNS or account-server behavior. Do not guess Nintendo service hostnames: confirm them from the actual game.

## Capture setup

1. Use a Wii U or a development setup you own and a dedicated test account.
2. Capture traffic at the router/AP or on a mirrored port so the capture includes the console's DNS queries and network connections. Record the Wii U's local IP in your private notes.
3. First record a baseline with the console using your usual working network path. Then capture Pretendo DNS and WiiLink DNS separately, changing only the DNS profile between captures.
4. Begin capture before launching Wii Sports Club; stop after its initial online error/success screen. Keep each capture short.

If using Wireshark and the console IP is known, this display filter can help isolate traffic:

```text
ip.addr == WII_U_LOCAL_IP && (dns || tcp || udp)
```

Replace `WII_U_LOCAL_IP` with the local address of your console. Do not assume NEX ports are the only network traffic; the account/service-discovery request may use HTTPS before the game opens UDP sessions.

## What to record

For each connection attempt, note:

- DNS query name and answer, including the resolver profile in use.
- Destination host/IP, transport, and destination port.
- The stage where the game stops: account sign-in, initial service discovery, club data, rankings, or matchmaking.
- Game region/title ID and console firmware, if known.
- Any visible error code.

A short table is sufficient for sharing findings:

| Time | Stage | Hostname (if visible) | Protocol/port | DNS answer | Result |
| --- | --- | --- | --- | --- | --- |
| `00:00` | startup | `confirmed.example` | `UDP/60014` | `192.0.2.10` | example only |

The sample row uses a reserved documentation address and is not a real service hostname or endpoint.

## Never publish

Do not upload captures with unredacted PNID credentials, passwords, authentication tickets, session tokens, cookies, or private keys. DNS and destination IPs can also reveal private network details; sanitize before sharing. Do not paste an authenticated packet payload just to identify a port.

## How this helps the implementation

The captures should answer two separate questions:

1. **Discovery:** What endpoint does the Wii U contact first, and what response supplies its NEX server URL?
2. **Routing:** After discovery, which hostname/address/port does the game actually attempt to contact?

Only after those are known should a project-controlled override, service-discovery adapter, or endpoint patch be implemented. See [SERVICE-DISCOVERY.md](SERVICE-DISCOVERY.md).
