# Browser-based cloud deployment (no Mac Terminal)

This guide prepares the server from a web control panel. It does not require you to use Terminal or log in to Linux. It **does require a cloud-hosting account** and hosting costs; this repo cannot create an account or authorize charges for you.

## What this deployment does

- Runs MongoDB plus the authentication and secure/game containers on an Ubuntu VPS.
- Uses a persistent Docker volume for the game DataStore and another for MongoDB, so recreating containers does not erase those data files.
- Uses the server's public IPv4 address in the service configuration.
- Automatically checks for newer container images every 10 minutes. The publishing workflow only publishes images after the GitHub build/integration workflow succeeds and only for the current `main` commit.
- Exposes only the intended game UDP ports in Docker. The account helper on TCP 8080 and MongoDB port 27017 are not published.

**Important:** this creates a running backend, not a completed Wii U revival. It cannot change Pretendo's or WiiLink's public DNS or hosted account server. Real console login still requires a compatible authorized account/service-discovery integration.

## 1. Wait for the tested images

1. Open [GitHub Actions](https://github.com/NVDEMU/pretendo-wii-sports-club/actions) and confirm the latest **WSC Build & Integration Tests (browser-run)** run is green.
2. Open [GitHub Packages for NVDEMU](https://github.com/users/NVDEMU?tab=packages). The publishing workflow runs automatically after a successful main-branch CI run.
3. Confirm these two packages exist:
   - `pretendo-wii-sports-club-authentication`
   - `pretendo-wii-sports-club-secure`
4. Set **both packages to Public** in each package's **Package settings → Change visibility**. The cloud host is configured to pull the public images without a private access token. Do not make unrelated packages public.

The publishing workflow is at [publish-containers.yml](../.github/workflows/publish-containers.yml). Its `main` tags change only after a successful CI build and integration test.

## 2. Create a VPS in the browser

This configuration is written for [DigitalOcean Droplets](https://www.digitalocean.com/pricing/droplets) using an Ubuntu 24.04 image. As listed today, a 2 GiB / 2 vCPU Basic Droplet is about **$12/month**; choose a smaller or larger plan only after considering MongoDB and runtime memory. Prices can change, and DigitalOcean may require account/payment verification. Creating a Droplet starts billing.

1. Open the [DigitalOcean control panel](https://cloud.digitalocean.com/).
2. Choose **Create → Droplets**.
3. Select **Ubuntu 24.04 LTS** and a Basic plan with at least 2 GiB RAM for the initial deployment.
4. Choose your desired datacenter region.
5. Choose an authentication option as required by the control panel. The startup script does not need you to connect over SSH after setup; keep credentials private and use the provider's secure account features.
6. Under **Additional Options**, enable **Startup scripts**.
7. Open [the cloud-init file in this repository](../deploy/digitalocean-cloud-init.yaml), copy its entire contents, and paste it into the Startup scripts field. Do not paste it into a terminal.
8. Create the Droplet and wait for its status to become active. First-boot setup can take several minutes.

Cloud-init runs automatically as root on the first boot; see [DigitalOcean's official user-data guide](https://docs.digitalocean.com/products/droplets/how-to/provide-user-data/).

## 3. Configure the provider firewall in the browser

In the DigitalOcean control panel, create a **Cloud Firewall** and attach it to your new Droplet. Permit these two inbound rules from all IPv4 clients:

| Protocol | Port | Purpose |
| --- | ---: | --- |
| UDP | 60014 | NEX authentication |
| UDP | 60015 | NEX secure/game service |

Leave inbound TCP 8080 and TCP 27017 closed. The account bridge and database are private container-network services, not public endpoints. Keep only the access rules you actually need. DigitalOcean Cloud Firewalls deny traffic that has no matching allow rule; see [the official firewall docs](https://docs.digitalocean.com/products/networking/firewalls/how-to/configure-rules/).

## 4. Check bootstrap status without a terminal

From the Droplet page, use the provider's browser-based Console if you need to inspect the startup log. The bootstrap writes `/var/log/wsc-bootstrap.log` and cloud-init writes `/var/log/cloud-init-output.log`. A message saying the GHCR image cannot be pulled usually means the image package is still private or the publishing workflow has not finished.

The containers use a private Docker network. The helper port 8080 is intentionally not shown as a public port.

## 5. What remains after deployment

Once the Droplet is active and both UDP firewall rules are in place, the hosting portion is ready for external packet-level checks. However:

1. The account-server route for game-server ID `1012F100` still needs to be configured on an account server you operate or are explicitly authorized to modify.
2. Public Pretendo/WiiLink DNS must route the appropriate service-discovery path, or a separately authorized client/service configuration must do so.
3. Actual Wii U login, rankings, matchmaking, and matches must be tested independently. A running VPS or passing CI is not proof of those features.

Do not redirect unrelated Nintendo hostnames, share account tokens, or expose the private bridge. Stop/delete the Droplet from the provider dashboard if you do not want to keep paying for it.
