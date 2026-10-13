import json
import pathlib
import re
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[2]
PATCH = ROOT / "integrations" / "pretendo-account" / "provider-route.patch"
ENDPOINTS = ROOT / "config" / "wii-sports-club-endpoints.json"


class AccountRoutePatchSafetyTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.patch = PATCH.read_text(encoding="utf-8")
        cls.endpoints = json.loads(ENDPOINTS.read_text(encoding="utf-8"))

    def test_route_is_wsc_only_and_uses_existing_nex_pid(self):
        self.assertIn("if (wscGameServerID?.toUpperCase() === '1012F100')", self.patch)
        self.assertIn("const gameServerID = getValueFromQueryString(request.query, 'game_server_id');", self.patch)
        self.assertIn("nexAccount.pid", self.patch)
        self.assertNotIn("NEXAccount.create", self.patch)
        self.assertNotIn("NEXAccount.update", self.patch)

    def test_title_allowlist_matches_repo_metadata(self):
        match = re.search(r"const allowedTitleIDs = new Set\(\[([\s\S]*?)\]\);", self.patch)
        self.assertIsNotNone(match)
        ids = re.findall(r"'([0-9A-Fa-f]{16})'", match.group(1))
        expected = (
            self.endpoints["title_ids"]["confirmed_by_pretendo_nex_viewer"]
            + self.endpoints["title_ids"]["additionally_reported_by_revivetendo_title_map"]
        )
        self.assertEqual(set(ids), set(expected))
        self.assertEqual(len(ids), len(set(ids)))

    def test_shared_hmac_and_direct_endpoint_response(self):
        self.assertIn("WSC_NEX_PASSWORD_SECRET", self.patch)
        self.assertIn("crypto.createHmac('sha256', wscPasswordSecret)", self.patch)
        self.assertIn(".update(String(nexAccount.pid)).digest('hex')", self.patch)
        self.assertIn("WSC_NEX_AUTH_HOST", self.patch)
        self.assertIn("WSC_NEX_AUTH_PORT", self.patch)
        self.assertIn("nex_password: wscNexPassword", self.patch)
        self.assertNotIn("got.post", self.patch)
        self.assertNotIn("WSC_INTERNAL_API_URL", self.patch)

    def test_token_registration_and_cache_headers(self):
        self.assertIn("NEXToken.create({", self.patch)
        self.assertIn("crypto.createHash('sha256').update(wscToken).digest('hex')", self.patch)
        self.assertIn("expires: new Date(issuedAt.getTime() + (3600 * 1000))", self.patch)
        self.assertIn("response.set('Cache-Control', 'no-store')", self.patch)
        self.assertIn("response.set('Pragma', 'no-cache')", self.patch)


if __name__ == "__main__":
    unittest.main()
