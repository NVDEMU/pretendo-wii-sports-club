#!/usr/bin/env python3
"""Query a chosen upstream DNS resolver and optionally verify the returned IPv4 address.

This is a preflight tool, not a DNS server. It does not modify Pretendo or WiiLink records.
Use only with a hostname known from a trusted configuration or a packet capture.
"""
from __future__ import annotations

import argparse
import ipaddress
import json
import secrets
import socket
import struct
import sys
from typing import Any

PROFILES = {
    "pretendo": "88.198.140.154",
    "wiilink": "167.235.229.36",
}


def encode_name(hostname: str) -> bytes:
    hostname = hostname.strip().rstrip(".")
    if not hostname:
        raise ValueError("hostname must not be empty")
    encoded = bytearray()
    for label in hostname.split("."):
        raw = label.encode("idna")
        if not raw or len(raw) > 63:
            raise ValueError(f"invalid DNS label in {hostname!r}")
        encoded.append(len(raw))
        encoded.extend(raw)
    encoded.append(0)
    return bytes(encoded)


def decode_name(packet: bytes, offset: int) -> tuple[str, int]:
    labels: list[str] = []
    next_offset: int | None = None
    visited: set[int] = set()

    while True:
        if offset >= len(packet):
            raise ValueError("truncated DNS name")
        size = packet[offset]
        if size & 0xC0 == 0xC0:
            if offset + 1 >= len(packet):
                raise ValueError("truncated DNS compression pointer")
            pointer = ((size & 0x3F) << 8) | packet[offset + 1]
            if pointer in visited:
                raise ValueError("DNS compression pointer loop")
            visited.add(pointer)
            if next_offset is None:
                next_offset = offset + 2
            offset = pointer
            continue
        if size & 0xC0:
            raise ValueError("unsupported DNS label encoding")
        offset += 1
        if size == 0:
            return ".".join(labels), next_offset if next_offset is not None else offset
        if offset + size > len(packet):
            raise ValueError("truncated DNS label")
        labels.append(packet[offset:offset + size].decode("idna"))
        offset += size


def build_query(hostname: str, query_id: int) -> bytes:
    question = encode_name(hostname) + struct.pack("!HH", 1, 1)  # A, IN
    header = struct.pack("!HHHHHH", query_id, 0x0100, 1, 0, 0, 0)
    return header + question


def parse_response(packet: bytes, query_id: int) -> list[dict[str, Any]]:
    if len(packet) < 12:
        raise ValueError("DNS response header is truncated")
    response_id, flags, qdcount, ancount, _, _ = struct.unpack_from("!HHHHHH", packet, 0)
    if response_id != query_id:
        raise ValueError("DNS response ID does not match query")
    if not flags & 0x8000:
        raise ValueError("packet is not marked as a DNS response")
    if flags & 0x0200:
        raise ValueError("truncated DNS response; retry using TCP or another resolver")
    rcode = flags & 0x000F
    if rcode:
        labels = {1: "FORMERR", 2: "SERVFAIL", 3: "NXDOMAIN", 4: "NOTIMP", 5: "REFUSED"}
        raise ValueError(f"DNS resolver returned {labels.get(rcode, f'rcode {rcode}')}")

    offset = 12
    for _ in range(qdcount):
        _, offset = decode_name(packet, offset)
        if offset + 4 > len(packet):
            raise ValueError("truncated DNS question")
        offset += 4

    records: list[dict[str, Any]] = []
    for _ in range(ancount):
        owner, offset = decode_name(packet, offset)
        if offset + 10 > len(packet):
            raise ValueError("truncated DNS answer")
        rrtype, rrclass, ttl, data_len = struct.unpack_from("!HHIH", packet, offset)
        data_offset = offset + 10
        end = data_offset + data_len
        if end > len(packet):
            raise ValueError("truncated DNS record data")

        if rrclass == 1 and rrtype == 1 and data_len == 4:
            address = socket.inet_ntop(socket.AF_INET, packet[data_offset:end])
            records.append({"name": owner, "type": "A", "ttl": ttl, "value": address})
        elif rrclass == 1 and rrtype == 5:
            target, _ = decode_name(packet, data_offset)
            records.append({"name": owner, "type": "CNAME", "ttl": ttl, "value": target})
        offset = end
    return records


def query_a(hostname: str, resolver: str, timeout: float) -> list[dict[str, Any]]:
    query_id = secrets.randbelow(65536)
    query = build_query(hostname, query_id)
    with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as sock:
        sock.settimeout(timeout)
        sock.sendto(query, (resolver, 53))
        packet, _ = sock.recvfrom(4096)
    return parse_response(packet, query_id)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--profile", choices=sorted(PROFILES), required=True,
                        help="public upstream resolver to query")
    parser.add_argument("--hostname", required=True,
                        help="verified hostname from configuration or captured traffic")
    parser.add_argument("--expected-ip",
                        help="expected IPv4 answer; exit non-zero if the resolver does not return it")
    parser.add_argument("--resolver",
                        help="override the documented resolver address for controlled testing")
    parser.add_argument("--timeout", type=float, default=3.0,
                        help="DNS UDP timeout in seconds (default: 3)")
    args = parser.parse_args()

    resolver = args.resolver or PROFILES[args.profile]
    try:
        ipaddress.IPv4Address(resolver)
        expected_ip = str(ipaddress.IPv4Address(args.expected_ip)) if args.expected_ip else None
        if args.timeout <= 0:
            raise ValueError("timeout must be greater than zero")
        records = query_a(args.hostname, resolver, args.timeout)
    except (OSError, ValueError) as exc:
        print(json.dumps({
            "profile": args.profile,
            "resolver": resolver,
            "hostname": args.hostname,
            "error": str(exc),
        }, indent=2))
        return 2

    addresses = [record["value"] for record in records if record["type"] == "A"]
    result = {
        "profile": args.profile,
        "resolver": resolver,
        "hostname": args.hostname.rstrip("."),
        "records": records,
        "expected_ip": expected_ip,
        "expected_ip_matches": expected_ip in addresses if expected_ip else None,
        "warning": "DNS resolution alone does not verify the Wii U account-server service-discovery response or UDP reachability.",
    }
    print(json.dumps(result, indent=2))
    if expected_ip and expected_ip not in addresses:
        return 1
    if not records:
        print("No A/CNAME answer records returned.", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
