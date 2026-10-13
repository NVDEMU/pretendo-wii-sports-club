# Browser-only quick start (no Mac Terminal required)

You can check this project without opening the Mac Terminal app. GitHub Actions builds the source and runs the integration smoke test on GitHub-hosted runners.

## Run the tests from the GitHub website

1. Open the [Actions page](https://github.com/NVDEMU/pretendo-wii-sports-club/actions).
2. Select **WSC Build & Integration Tests (browser-run)** in the workflow list.
3. Select **Run workflow** on the right.
4. Keep the branch set to `main`, then select the green **Run workflow** button.
5. Open the new run. Wait for all jobs to finish, then check that each job has a green check mark. Open any failed job to see the error log.

The workflow checks both Go services, runs the Python DNS/configuration tests, validates the Compose configuration, builds both containers, and starts a disposable MongoDB-backed stack to test the private NEX-token bridge. It verifies invalid-token rejection, XML response shape, cache headers, stable per-PID password, and a rotating token.

The same checks run automatically whenever a commit is pushed to the default branch. You do not need to open Terminal, install Go or Python, or run a command on your Mac just to test a repository change.

## Make small changes in the browser

1. Open the [repository](https://github.com/NVDEMU/pretendo-wii-sports-club).
2. Browse to a file.
3. Use the pencil **Edit this file** button where available, or choose **Add file** to create a file.
4. Make your edit in GitHub's web editor.
5. Select **Commit changes** and commit to `main` (or create a branch/pull request if you prefer).
6. Open [Actions](https://github.com/NVDEMU/pretendo-wii-sports-club/actions) to see the checks automatically triggered by that commit.

Do not enter real account passwords, console credentials, access tokens, or secret values in source files or commits. Keep secrets out of issues and public logs.

## What a green run means

A green run means the code/build/configuration checks and local container smoke test passed in GitHub's temporary test environment. It does **not** mean the server has been deployed publicly, the Wii U's account/service-discovery route is configured, or a real Wii U login/match succeeded.

Actual online play still requires a public host that supports inbound UDP 60014 and 60015 and an authorized account/service-discovery route. GitHub Actions is for build and integration testing, not long-running public UDP hosting.
