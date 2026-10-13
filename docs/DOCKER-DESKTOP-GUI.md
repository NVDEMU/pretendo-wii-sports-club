# Run the WSC server in Docker Desktop without Terminal

This is the graphical, school-Mac-friendly way to start a **local testing copy** of the WSC server. It uses one container with MongoDB and both WSC servers inside it, so you do not need to configure a three-container network. The aim is to keep testing simple while we finish the account/service-discovery integration.

**This only makes a local server process run. It does not configure Nintendo/Pretendo/WiiLink account routing or guarantee that a Wii U can log in.** The server advertises `localhost` by default, which is appropriate for startup smoke testing only. Do not expose it as a public service.

## A. Get a runnable image

The WSC image has been built and published to GitHub Container Registry (GHCR), but Docker Desktop's documented GUI pull workflow is for Docker Hub. Since this guide is for users who cannot use Terminal, the **browser-only route is to publish a copy to your Docker Hub account**. This does not host your game server on Docker Hub; it only stores the downloadable image.

You do not need a cloud VM. Do not follow these steps on a school-managed account/device if your school does not allow creating registry accounts, publishing public images, or adding GitHub Actions secrets.

1. Sign up for a free [Docker Personal account](https://hub.docker.com/signup). Docker Personal includes public repositories at no cost; see [current plan details](https://www.docker.com/products/personal/).
2. In [Docker Hub](https://hub.docker.com/), choose **My Hub → Repositories → Create repository**.
3. Create a repository named exactly `pretendo-wii-sports-club-desktop` under your own namespace. Choose **Public** so Docker Desktop can pull it anonymously. Public images can be downloaded by anyone.
4. In Docker Hub, open your avatar → **Account settings → Personal access tokens → Generate new token**. Give it a label such as `WSC GitHub Actions` and choose **Read & Write** access (not Delete). Save the token somewhere safe; Docker shows it only once. Treat it like your password. See [Docker's token guide](https://docs.docker.com/security/access-tokens/personal-access-tokens/).
5. In this GitHub repository, open **Settings → Secrets and variables → Actions → New repository secret**. Add these two secrets:
   - Name: `DOCKERHUB_USERNAME`; value: your Docker Hub username.
   - Name: `DOCKERHUB_TOKEN`; value: the token you just created (not your Docker password).
6. After both secrets are saved, open [GitHub Actions](../actions), choose **WSC Build & Integration Tests (browser-run)**, click **Run workflow**, keep the branch set to `main`, and confirm. Wait for all CI jobs to pass and for **Publish tested WSC containers** to finish. The publisher uploads the desktop image to Docker Hub automatically when both secrets are present. Confirm that the `main` tag appears in your Docker Hub repository. Do not manually publish an untested revision.

GitHub secrets are entered directly in GitHub's settings page and are not committed to the repository. Never paste the token into source code, an issue, a chat, or a public log. Official instructions: [GitHub Actions repository secrets](https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/use-secrets).

## B. Get the image into Docker Desktop

1. Open Docker Desktop and make sure its engine is running.
2. Select **Images** and choose **Search images to run** (or use the Docker Desktop search control).
3. Search for `YOUR_DOCKERHUB_USERNAME/pretendo-wii-sports-club-desktop`, using the Docker Hub username whose credentials you stored in GitHub.
4. Select the `main` tag and click **Pull**. If it is not available, check that the GitHub Actions secrets are correct and that **Publish tested WSC containers** completed successfully.

Docker Desktop's image details show the image size and provide actions for removing it later. See [Docker's Images view guide](https://docs.docker.com/desktop/use-desktop/images/).

## C. Run the container in the GUI

1. In Docker Desktop's **Images** view, find the pulled `pretendo-wii-sports-club-desktop` image and click **Run**.
2. Expand **Optional settings**.
3. Set the container name to `wsc-local`.
4. In **Ports**, publish UDP port `60014` from the container to host port `60014`, and UDP port `60015` to host port `60015`. Keep the UDP protocol. Do not publish TCP `8080` or MongoDB `27017`.
5. In **Volumes**, the image declares `/data` for persistent database and DataStore files. Keep the volume mounted at `/data`. If the UI lets you choose a named volume, name it `wsc-local-data`; otherwise Docker will create an anonymous volume for the declared mount. Keep the same container/volume if you want to retain local test data.
6. In **Environment variables**, add:
   - `KERBEROS_PASSWORD`: choose a long random value for local testing. Do not reuse an account password.
   - `SECURE_SERVER_LOCATION`: `localhost` for local startup testing.
   - `SECURE_SERVER_PORT`: `60015` (usually already set as the default).
7. Click **Run**.

The image starts a local MongoDB instance, then launches the authentication server and secure/game server. No MongoDB port is exposed on the host.

**This is not a website.** The bundle serves the game over UDP on ports 60014 and 60015, so opening a port as `http://localhost:60014` or `http://localhost:60015` can produce an HTTP 404 or no response; that alone does not mean the container failed. Verify it in Docker Desktop → **Containers** → `wsc-local` → **Logs**. If Docker Desktop itself shows an HTTP 404 popup immediately after you click **Run** and no container appears, that is a different Docker Desktop error; capture the complete popup and check whether `wsc-local` appears in Containers.

## D. Check the result

1. Open **Containers** in Docker Desktop.
2. Select `wsc-local`.
3. Check that its status is **Running** / **Healthy**.
4. Open **Logs** and look for:
   - `MongoDB is ready`
   - `starting the WSC secure/game server`
   - `Starting the WSC NEX authentication server`

If it exits, check the same Logs view for the first error. Do not share environment-variable values or passwords in screenshots.

To stop the local server, use the Stop button in Docker Desktop. The Mac must stay awake with Docker Desktop running for the container to stay available.

## Storage use and cleanup

**Yes, this uses storage on the Mac.** Docker stores downloaded images, containers, and their data inside Docker Desktop's disk image. The exact space used can vary by build; after pulling, check the image's **Size** column or open its details in **Images**. Docker Desktop's **Settings → Resources → Advanced** also shows actual disk usage and the disk-image location. See [Docker's macOS storage guidance](https://docs.docker.com/desktop/troubleshoot-and-support/faqs/macfaqs/).

To reclaim space using the GUI, stop and remove the `wsc-local` container, then delete its desktop image from **Images**. If you also want to erase the saved local database and game data, remove its associated volume from **Volumes**; deleting that volume permanently deletes the test data. Removing the container/image alone might leave a persistent volume behind.

## What local hosting proves

A healthy container proves the processes can start together on your Mac. It does not verify console login, matchmaking, or online play. The default `localhost` advertised host is not reachable as a server address from a Wii U. A console test would require a suitable Mac LAN address plus an account/service-discovery route and network access; a school network may not permit inbound game traffic. Do not bypass school restrictions or alter the managed network settings.

If your school does not permit publishing Docker images to a public Docker Hub account or adding GitHub Actions secrets, do not do those steps on the school account/device. Continue with the GitHub-hosted CI path instead.


## Existing Pretendo PNIDs

You do not need to create a second player account. The public account route and WSC auth server must share `WSC_NEX_PASSWORD_SECRET`; it is a server-side secret, not a PNID password. Do not invent a different secret on the Mac and expect Pretendo's route to match. Leave it unset for a local startup test; that uses the local database fallback and does not enable public Pretendo login.
