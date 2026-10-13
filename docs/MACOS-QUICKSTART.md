# macOS and browser-based development

You do **not** need to install Linux to work on this project. There are two development options.

## Option A: run it on your Mac

### 1. Install Docker Desktop

Download and install the Mac version of [Docker Desktop](https://docs.docker.com/desktop/setup/install/mac-install/). Choose the installer for your Mac's processor (Apple silicon or Intel), move Docker to Applications if prompted, and open Docker.app. Wait for Docker Desktop to finish starting.

Docker Desktop supplies Docker Engine, the Docker CLI, and Docker Compose. The containers themselves use Linux images, but Docker Desktop manages the Linux VM for you; you do not need to install or administer Linux.

### 2. Download this repository

Open Terminal and paste:

```sh
mkdir -p ~/Projects
cd ~/Projects
git clone https://github.com/NVDEMU/pretendo-wii-sports-club.git
cd pretendo-wii-sports-club
```

If you already downloaded the repository, use `cd` to enter its folder instead of cloning it again.

### 3. Start the local stack and run the smoke test

Run this from the repository folder:

```sh
bash scripts/macos-local-start.sh
```

The script creates a local `.env` with random development-only secrets if one does not exist, builds the containers, starts MongoDB and both server processes, and runs the same internal-bridge/MongoDB smoke test used by CI. It will not overwrite an existing `.env`. The file contains secrets and is git-ignored; do not upload it or paste its contents into an issue.

Useful commands:

```sh
docker compose -p wsc-local ps
docker compose -p wsc-local logs --tail=100 wsc-authentication wsc-secure
docker compose -p wsc-local down
```

If the script reports that Docker is not running, open Docker.app and try again. If the smoke test fails, copy the **error output only** (not your `.env`) into an issue so it can be debugged.

## Option B: use GitHub Codespaces in a browser

This repository now includes a `.devcontainer/devcontainer.json` with Go, Python, Docker-in-Docker, and recommended editor extensions.

1. Open the [repository](https://github.com/NVDEMU/pretendo-wii-sports-club).
2. Click **Code**, open the **Codespaces** tab, then choose **Create codespace on main**.
3. Wait for the environment setup to finish. The existing Python DNS tests run during setup.
4. Use the Codespaces terminal for development. For the full Compose smoke test, run `bash scripts/macos-local-start.sh`; Codespaces must have a working Docker daemon for this to run.

GitHub may require an account with Codespaces access and applicable usage allowance.

## What local testing proves — and what it does not

A successful local smoke test verifies that the authentication container can use MongoDB, rejects an invalid bridge secret, returns the expected XML fields and no-cache headers, keeps the NEX password stable per PID, and rotates the separate token across requests. CI independently runs the same smoke test on each relevant push.

Local testing does **not** prove that a Wii U can connect over the Internet. The generated configuration advertises `localhost`, which is deliberately local-only. Do not put it into a public DNS profile or try it as the endpoint on a real console.

## The remaining step for actual online Wii U tests

The server needs a publicly reachable host that allows inbound **UDP 60014 and UDP 60015**, plus a compatible account/service-discovery route that advertises the authentication endpoint. The project's private HTTP bridge on TCP 8080 must remain private. DNS resolver lookups alone do not configure that route.

You can prepare and validate the code on your Mac or in Codespaces first. A public server and authorization to configure its account/service-discovery path are separate requirements; installing Linux locally is not required for either.
