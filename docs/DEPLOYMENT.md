# Deployment and live verification

This guide runs the community NEX services on a Linux host. It does **not** configure Pretendo's or WiiLink's hosted DNS/account services; it is for a server operator who controls the host and an account-service route they are authorized to modify.

## 1. Prepare a Linux host

Use a public VPS or server with Docker Engine and the Docker Compose plugin. Configure DNS A records for the public NEX authentication and secure-service hostnames to resolve to this host. If you use one hostname for both services, it must resolve to this host as well.

Open inbound **UDP 60014** and **UDP 60015** in both the host firewall and the provider's network firewall. TCP 8080 is a private helper port and must not be publicly exposed.

## 2. Configure the service

From the repository root:

```sh
cp .env.example .env
openssl rand -hex 32
```

Use the generated random value as `KERBEROS_PASSWORD`. Generate another independent value with `openssl rand -hex 32` for `WSC_INTERNAL_API_TOKEN`. Set:

- `SECURE_SERVER_LOCATION`: public host/IP that resolves to the NEX secure-service endpoint.
- `SECURE_SERVER_PORT=60015`.
- `NEX_AUTH_PUBLIC_HOST`: public host/IP that resolves to the NEX authentication endpoint.
- `NEX_AUTH_PUBLIC_PORT=60014`.
- `WSC_INTERNAL_API_TOKEN`: long random bearer secret; share it only with the trusted account-server process that calls the internal bridge.

Start the services:

```sh
docker compose up --build -d
docker compose ps
docker compose logs --tail=100 wsc-authentication wsc-secure
```

Check that both NEX processes stay running and MongoDB reports healthy. Do not commit `.env`, paste credentials into issues, or expose TCP 8080.

## 3. Connect an authorized account service

Review [the self-hosted account-server integration](../integrations/pretendo-account/README.md). Apply the route patch only to an account server you operate or are explicitly authorized to modify. The account-server process must be able to reach:

```text
http://wsc-authentication:8080/internal/wsc/nex-token
```

That container DNS name works only when both containers share a Docker network. Otherwise, configure `WSC_INTERNAL_API_URL` to a private address routable only from the trusted account server, and keep port 8080 blocked from the public network.

Set the same `WSC_INTERNAL_API_TOKEN` in both services. Do **not** redirect every Nintendo hostname to this server or modify another provider's public resolver; only configure an account/service-discovery route you are authorized to administer.

## 4. Run local/container validation

CI now runs an end-to-end smoke test against the private bridge and a real MongoDB container. It checks invalid bearer rejection, the XML response fields, cache headers, stable NEX password for repeated calls to one PID, and a new token on each request.

Run that check in GitHub Actions at [the latest workflow runs](../actions). It does not require your Mac to build the source locally.

## 5. Verify actual Wii U traffic

Only after the account-service request has been routed to a compatible response:

1. Use a dedicated test account and a Wii U you control.
2. Capture a startup attempt and redact access tokens, NEX tokens, passwords, and device certificates before sharing anything.
3. Confirm the account response advertises the **NEX authentication host/UDP port 60014**.
4. Confirm the NEX authentication response advertises the **secure server/UDP port 60015**.
5. Test login first, then profile/club data, rankings, matchmaking, and each sport separately.
6. Repeat after switching DNS profiles. A successful DNS lookup alone is not proof that this endpoint was advertised or reached.

The repository smoke test only verifies the helper service and database integration. No end-to-end Wii U login or live online match should be considered verified until the above test succeeds on hardware.
