import json
import pathlib
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[2]
CONFIG = ROOT / "config" / "wii-sports-club-endpoints.json"


class EndpointReferenceTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.data = json.loads(CONFIG.read_text(encoding="utf-8"))

    def test_wsc_game_identity_is_present(self):
        self.assertEqual(self.data["game_server_id"], "1012F100")
        self.assertEqual(self.data["access_key"], "4d324052")

    def test_discovery_path_and_query_are_exact(self):
        discovery = self.data["service_discovery"]
        self.assertEqual(discovery["host"], "account.nintendo.net")
        self.assertEqual(discovery["path"], "/v1/api/provider/nex_token/@me")
        self.assertEqual(discovery["method"], "GET")
        self.assertEqual(discovery["query"]["game_server_id"], "1012F100")
        for field in ("host", "port", "pid", "nex_password", "token"):
            self.assertIn(field, discovery["response_fields"])

    def test_community_ports_are_not_the_old_nintendo_endpoint(self):
        servers = self.data["nex_server_configuration"]
        self.assertEqual(servers["community_auth_endpoint"]["port"], 60014)
        self.assertEqual(servers["community_secure_endpoint"]["port"], 60015)
        historical = self.data["historical_nintendo_endpoint"]
        self.assertEqual(historical["ipv4"], "34.208.166.202")
        self.assertEqual(historical["port"], 43480)
        self.assertNotEqual(
            servers["community_auth_endpoint"]["port"],
            historical["port"],
        )
        self.assertIn("Historical endpoint data only", historical["status"])

    def test_known_title_ids_are_well_formed(self):
        title_ids = self.data["title_ids"]
        all_ids = (
            title_ids["confirmed_by_pretendo_nex_viewer"]
            + title_ids["additionally_reported_by_revivetendo_title_map"]
        )
        self.assertIn("000500001012F100", all_ids)
        self.assertTrue(all(len(title_id) == 16 for title_id in all_ids))
        self.assertTrue(all(set(title_id) <= set("0123456789ABCDEF") for title_id in all_ids))

    def test_both_dns_resolvers_are_documented_as_resolvers(self):
        resolvers = self.data["dns_resolvers"]
        self.assertEqual(resolvers["pretendo_sssl"]["ipv4"], "88.198.140.154")
        self.assertEqual(resolvers["wiilink"]["ipv4"], "167.235.229.36")
        self.assertIn("DNS settings alone", resolvers["constraint"])

    def test_live_service_discovery_is_not_claimed_as_implemented(self):
        status = self.data["service_discovery"]["status"].lower()
        self.assertIn("does not currently implement", status)


if __name__ == "__main__":
    unittest.main()
