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

    def test_patch_targets_only_wsc_game_server_id(self):
        self.assertIn("if (wscGameServerID?.toUpperCase() === '1012F100')", self.patch)
        self.assertIn("const gameServerID = getValueFromQueryString(request.query, 'game_server_id');", self.patch)
        # The generic route remains after the WSC-only branch.
        self.assertIn("const nexAccount = await NEXAccount.findOne({", self.patch)
        self.assertLess(self.patch.index("const nexAccount = await NEXAccount.findOne({"),
                        self.patch.index("const wscGameServerID = getValueFromQueryString"))

    def test_patch_allowlist_matches_documented_title_id_candidates(self):
        match = re.search(r"const allowedTitleIDs = new Set\(\[([\s\S]*?)\]\);", self.patch)
        self.assertIsNotNone(match, "patch needs an explicit fail-closed title ID allowlist")
        ids = re.findall(r"'([0-9A-Fa-f]{16})'", match.group(1))
        expected = (
            self.endpoints["title_ids"]["confirmed_by_pretendo_nex_viewer"]
            + self.endpoints["title_ids"]["additionally_reported_by_revivetendo_title_map"]
        )
        self.assertEqual(set(ids), set(expected))
        self.assertEqual(len(ids), len(set(ids)))

    def test_bridge_uses_the_authenticated_users_existing_nex_pid(self):
        self.assertIn("json: { pid: nexAccount.pid, token: wscToken }", self.patch)
        self.assertNotIn("json: { pid: pnid.pid }", self.patch)
        self.assertIn("await NEXToken.create({", self.patch)
        self.assertIn("crypto.createHash('sha256').update(wscToken).digest('hex')", self.patch)
        self.assertIn("system_type: SystemType.WUP", self.patch)
        self.assertNotIn("request.body.pid", self.patch)
        self.assertNotIn("request.query.pid", self.patch)
        self.assertIn("process.env.WSC_INTERNAL_API_URL", self.patch)
        self.assertIn("process.env.WSC_INTERNAL_API_TOKEN", self.patch)
        self.assertIn("bridgeToken.length < 32", self.patch)

    def test_session_xml_is_not_cacheable(self):
        self.assertIn("response.set('Cache-Control', 'no-store')", self.patch)
        self.assertIn("response.set('Pragma', 'no-cache')", self.patch)
        self.assertIn("response.status(200).type('application/xml').send(tokenXML)", self.patch)

    def test_patch_notes_its_authorization_boundary(self):
        self.assertIn("account server you operate or are authorized to modify", self.patch)
        self.assertIn("never expose its TCP port publicly", self.patch)


if __name__ == "__main__":
    unittest.main()
