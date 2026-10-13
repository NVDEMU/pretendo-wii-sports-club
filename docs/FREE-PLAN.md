# Zero-cost development plan

**You do not need to pay for a server to keep making progress.** Do not create a paid VPS or enter billing details just to continue development.

## What can be done for $0

- Edit source, YAML, tests, and documentation in the GitHub web editor.
- Run the existing build and integration checks with GitHub Actions. For a public repository, standard GitHub-hosted Actions runners are free under GitHub's current policy: [GitHub Actions billing](https://docs.github.com/en/billing/concepts/product-billing/github-actions).
- Publish public container images with GitHub Container Registry; GitHub currently lists public container registry image storage/bandwidth as free: [GitHub Packages billing](https://docs.github.com/en/billing/concepts/product-billing/github-packages). This policy may change; review the current terms if usage grows.
- Expand offline tests with deterministic fake account-server responses, DNS configuration tests, XML validation, datastore persistence tests, and regressions based on already-redacted packet captures.
- Use GitHub Issues and pull requests to get review from other community developers, without sharing passwords, tokens, account credentials, or device certificates.

## Run checks entirely in your browser

1. Open [GitHub Actions](https://github.com/NVDEMU/pretendo-wii-sports-club/actions).
2. Select **WSC Build & Integration Tests (browser-run)**.
3. Choose **Run workflow**, use branch `main`, and confirm.
4. Review the job results. Green checks confirm the code/build/integration tests passed in GitHub's temporary environment.

You can also edit files by opening them in GitHub, choosing the pencil icon when available, and committing your changes. Pushes to the default branch trigger the tests automatically. No local Terminal, Linux, or cloud account is required.

## What cannot be proven without an actual reachable server

A GitHub Actions runner is temporary and does not provide a permanent public UDP endpoint for Wii U clients. Passing tests does not mean the console can log in or play a match.

Public Wii U verification still needs a server that can run continuously and accept inbound UDP 60014 and 60015, plus a compatible account/service-discovery route operated by us or configured by a provider with authorization. If you cannot pay for hosting, leave deployment on hold until a genuinely free host is available or a community member voluntarily offers hosting. Never assume a provider's free trial is permanent, and never create a resource unless the control panel clearly shows it is free.

## Optional cloud guide

The [browser cloud deployment guide](BROWSER-DEPLOYMENT.md) describes a paid VPS option. **Skip that guide for now if your budget is $0.** No paid resource is required to continue coding and testing this repo.
