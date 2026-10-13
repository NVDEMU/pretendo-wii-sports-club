#!/usr/bin/env python3
"""End-to-end smoke test for the private NEX-token bridge in Docker Compose."""
from __future__ import annotations

import argparse
import json
import os
import sys
import time
import urllib.error
import urllib.request
import xml.etree.ElementTree as ET


def request(url: str, method: str, token: str | None = None, payload: dict | None = None):
    headers = {}
    body = None
    if token is not None:
        headers["Authorization"] = f"Bearer {token}"
    if payload is not None:
        headers["Content-Type"] = "application/json"
        body = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(url, data=body, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=4) as response:
            return response.status, dict(response.headers), response.read()
    except urllib.error.HTTPError as exc:
        return exc.code, dict(exc.headers), exc.read()


def parse_token_xml(body: bytes) -> dict[str, str]:
    root = ET.fromstring(body)
    if root.tag != "nex_token":
        raise AssertionError(f"expected <nex_token>, got <{root.tag}>")
    result = {child.tag: (child.text or "") for child in root}
    required = {"host", "port", "pid", "nex_password", "token"}
    missing = required - result.keys()
    if missing:
        raise AssertionError(f"missing XML fields: {', '.join(sorted(missing))}")
    return result


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--url", default="http://wsc-authentication:8080/internal/wsc/nex-token")
    parser.add_argument("--host", default="wsc-smoke.example.invalid")
    parser.add_argument("--port", type=int, default=60014)
    parser.add_argument("--pid", type=int, default=2147483000)
    args = parser.parse_args()
    token = os.environ.get("WSC_INTERNAL_API_TOKEN", "")
    if len(token) < 32:
        print("WSC_INTERNAL_API_TOKEN must be provided and at least 32 characters", file=sys.stderr)
        return 2

    # Wait for the authentication service and MongoDB to become ready.
    deadline = time.monotonic() + 40
    last_error = None
    while time.monotonic() < deadline:
        try:
            status, _, _ = request(args.url, "POST", token="invalid", payload={"pid": args.pid})
            if status == 401:
                break
            last_error = f"expected unauthorized readiness response, got HTTP {status}"
        except (OSError, urllib.error.URLError) as exc:
            last_error = str(exc)
        time.sleep(1)
    else:
        print(f"private bridge did not become ready: {last_error}", file=sys.stderr)
        return 1

    status, _, _ = request(args.url, "POST", token="invalid", payload={"pid": args.pid})
    if status != 401:
        raise AssertionError(f"bad bearer token should return HTTP 401, got {status}")

    first_status, first_headers, first_body = request(
        args.url, "POST", token=token, payload={"pid": args.pid}
    )
    if first_status != 200:
        raise AssertionError(f"valid request returned HTTP {first_status}: {first_body[:300]!r}")
    if first_headers.get("Cache-Control", "").lower() != "no-store":
        raise AssertionError("token response must carry Cache-Control: no-store")
    first = parse_token_xml(first_body)

    second_status, _, second_body = request(
        args.url, "POST", token=token, payload={"pid": args.pid}
    )
    if second_status != 200:
        raise AssertionError(f"second valid request returned HTTP {second_status}")
    second = parse_token_xml(second_body)

    for data in (first, second):
        if data["host"] != args.host:
            raise AssertionError(f"unexpected host: {data['host']!r}")
        if data["port"] != str(args.port):
            raise AssertionError(f"unexpected port: {data['port']!r}")
        if data["pid"] != str(args.pid):
            raise AssertionError(f"unexpected PID: {data['pid']!r}")
        if not data["nex_password"]:
            raise AssertionError("nex_password must not be empty")
        if not data["token"]:
            raise AssertionError("token must not be empty")
        if data["token"] == data["nex_password"]:
            raise AssertionError("token and nex_password must be distinct")

    if first["nex_password"] != second["nex_password"]:
        raise AssertionError("NEX account password changed between requests for the same PID")
    if first["token"] == second["token"]:
        raise AssertionError("token should be freshly generated for each request")

    print("PASS: private bridge auth, XML schema, no-store, stable PID password, and rotating token")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (AssertionError, ET.ParseError, OSError, urllib.error.URLError) as exc:
        print(f"FAIL: {exc}", file=sys.stderr)
        raise SystemExit(1)
