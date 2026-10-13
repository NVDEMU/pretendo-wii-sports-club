# Run the WSC server in Docker Desktop without Terminal

This is the graphical, school-Mac-friendly way to start a **local testing copy** of the WSC server. No Terminal or local Linux installation is required. It uses one container with MongoDB and both WSC servers inside it, so you do not need to configure a three-container network.

**This only makes a local server process run. It does not configure Nintendo/Pretendo/WiiLink account routing or guarantee that a Wii U can log in.** The server advertises `localhost` by default, which is appropriate for startup smoke testing only. Do not expose it as a public service.

## A. One-time browser setup

Docker Desktop's Images search is designed around Docker Hub. To pull this project in the GUI without a command line, publish the bundle to your own free public Docker Hub repository.

1. Sign up for a free [Docker Personal account](https://hub.docker.com/signup). Docker Personal includes public repositories at no cost; see [current plan details](https://www.docker.com/products/personal/).
2. In [Docker Hub](https://hub.docker.com/), choose **My Hub → Repositories → Create repository**.
3. Create a repository named exactly `pretendo-wii-sports-club-desktop` under your own namespace. Choose **Public** so Docker Desktop can pull it anonymously. Public images can be downloaded by anyone.
4. In Docker Hub, open your avatar → **Account settings → Personal access tokens → Generate new token**. Give it a label such as `WSC GitHub Actions` and choose **Read & Write** access (not Delete). Save the token somewhere safe; Docker shows it only once. Treat it like your password. See [Docker's token guide](https://docs.docker.com/security/access-tokens/personal-access-tokens/).
5. In this GitHub repository, open **Settings → Secrets and variables → Actions → New repository secret**. Add these two secrets:
   - Name: `DOCKERHUB_USERNAME`; value: your Docker Hub username.
   - Name: `DOCKERHUB_TOKEN`; value: the token you just created (not your Docker password).
6. Open [GitHub Actions](../actions), choose **Publish tested WSC containers**, click **Run workflow**, keep the branch set to `main`, and confirm. This republishes the latest `main` revision and uploads the GUI bundle to your Docker Hub repository when both secrets are present. Confirm that `main` and a commit-SHA tag appear on Docker Hub. The automated publisher also runs after a successful main-branch CI result.

GitHub secrets are entered directly in GitHub's settings page and are not committed to the repository. Never paste the token into source code, an issue, a chat, or a public log. Official instructions: [GitHub Actions repository secrets](https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/use-secrets).

## B. Pull the image through Docker Desktop

1. Open Docker Desktop and make sure it says the Docker Engine is running.
2. Select **Images** from the left navigation.
3. Choose **Search images to run** or click the global search bar.
4. Search for your full image name: `YOUR_DOCKERHUB_USERNAME/pretendo-wii-sports-club-desktop`, replacing the placeholder with your Docker Hub username.
5. Select the repository, choose the `main` tag, then click **Pull**. If it does not appear in search immediately, refresh Docker Hub or open your public repository page to verify the tag is present first.

Docker's current UI guide covers searching, pulling, and running images: [Explore the Images view](https://docs.docker.com/desktop/use-desktop/images/).

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

## What local hosting proves

A healthy container proves the processes can start together on your Mac. It does not verify console login, matchmaking, or online play. The default `localhost` advertised host is not reachable as a server address from a Wii U. A console test would require a suitable Mac LAN address plus an account/service-discovery route and network access; a school network may not permit inbound game traffic. Do not bypass school restrictions or alter the managed network settings.

If your school does not permit publishing Docker images to a public Docker Hub account or adding GitHub Actions secrets, do not do those steps on the school account/device. Continue with the GitHub-hosted CI path instead.
