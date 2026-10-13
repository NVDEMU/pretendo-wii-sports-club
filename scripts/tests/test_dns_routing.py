import pathlib
import socket
import struct
import sys
import unittest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parents[1]))
import check_dns_routing as dns


def make_a_response(query_id: int, ip: str = "203.0.113.10") -> bytes:
    qname = dns.encode_name("server.example.test")
    question = qname + struct.pack("!HH", 1, 1)
    answer = b"\xc0\x0c" + struct.pack("!HHIH", 1, 1, 60, 4) + socket.inet_aton(ip)
    header = struct.pack("!HHHHHH", query_id, 0x8180, 1, 1, 0, 0)
    return header + question + answer


class DNSRoutingTests(unittest.TestCase):
    def test_encode_name_root_terminator(self):
        self.assertTrue(dns.encode_name("server.example.test").endswith(b"\x00"))

    def test_parse_a_answer(self):
        response = make_a_response(0x1234)
        parsed = dns.parse_response(response, 0x1234)
        self.assertEqual(len(parsed), 1)
        self.assertEqual(parsed[0]["type"], "A")
        self.assertEqual(parsed[0]["name"], "server.example.test")
        self.assertEqual(parsed[0]["value"], "203.0.113.10")
        self.assertEqual(parsed[0]["ttl"], 60)

    def test_mismatched_query_id_is_rejected(self):
        with self.assertRaisesRegex(ValueError, "ID"):
            dns.parse_response(make_a_response(0x1234), 0x4321)

    def test_nxdomain_is_reported(self):
        qname = dns.encode_name("missing.example.test")
        question = qname + struct.pack("!HH", 1, 1)
        header = struct.pack("!HHHHHH", 0x1234, 0x8183, 1, 0, 0, 0)
        with self.assertRaisesRegex(ValueError, "NXDOMAIN"):
            dns.parse_response(header + question, 0x1234)

    def test_truncated_packet_is_rejected(self):
        with self.assertRaisesRegex(ValueError, "truncated"):
            dns.parse_response(b"\x00\x01", 1)


if __name__ == "__main__":
    unittest.main()
