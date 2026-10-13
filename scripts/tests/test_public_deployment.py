import pathlib
import unittest
import yaml

ROOT = pathlib.Path(__file__).resolve().parents[2]


class PublicDeploymentTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.compose = yaml.safe_load((ROOT / "docker-compose.cloud.yml").read_text(encoding="utf-8"))
        cls.bootstrap = (ROOT / "deploy/oci-wsc-bootstrap.sh").read_text(encoding="utf-8")
        cls.guide = (ROOT / "docs/OCI-WSC-HOSTING.md").read_text(encoding="utf-8")
        cls.route = (ROOT / "integrations/pretendo-account/provider-route.patch").read_text(encoding="utf-8")

    def test_only_nex_udp_ports_are_published_by_game_services(self):
        auth = self.compose["services"]["wsc-authentication"]
        secure = self.compose["services"]["wsc-secure"]
        self.assertEqual(auth["ports"], ["60014:60014/udp"])
        self.assertEqual(secure["ports"], ["60015:60015/udp"])
        self.assertNotIn("8080:8080", str(self.compose))
        self.assertNotIn("27017:27017", str(self.compose))

    def test_shared_hmac_secret_is_required_for_public_runtime(self):
        auth = self.compose["services"]["wsc-authentication"]
        self.assertIn("WSC_NEX_PASSWORD_SECRET", auth["environment"])
        self.assertIn("WSC_NEX_PASSWORD_SECRET", self.bootstrap)
        self.assertIn("crypto.createHmac('sha256', wscPasswordSecret)", self.route)
        # Keep this a user-facing documentation check, not a dependency on
        # the internal Go function's symbol name.
        self.assertIn("derived WSC password", self.guide)

    def test_existing_pretendo_accounts_are_the_intended_path(self):
        self.assertIn("existing Pretendo PNIDs", self.guide)
        self.assertIn("WSC-only account-service route", self.guide)
        self.assertIn("does not create a replacement account system", self.guide)
        self.assertNotIn("WSC_BRIDGE_DOMAIN", self.bootstrap)
        self.assertIn("does not require a domain", self.guide)


if __name__ == "__main__":
    unittest.main()
