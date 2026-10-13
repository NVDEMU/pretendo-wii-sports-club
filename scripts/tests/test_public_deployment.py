import pathlib
import unittest
import yaml

ROOT = pathlib.Path(__file__).resolve().parents[2]


class PublicDeploymentTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.compose = yaml.safe_load((ROOT / "docker-compose.public.yml").read_text(encoding="utf-8"))
        cls.caddyfile = (ROOT / "deploy/Caddyfile").read_text(encoding="utf-8")
        cls.bootstrap = (ROOT / "deploy/oci-wsc-bootstrap.sh").read_text(encoding="utf-8")
        cls.guide = (ROOT / "docs/OCI-WSC-HOSTING.md").read_text(encoding="utf-8")

    def test_tls_bridge_proxy_is_limited_to_token_route(self):
        self.assertIn("@bridge path /internal/wsc/nex-token", self.caddyfile)
        self.assertIn("reverse_proxy @bridge wsc-authentication:8080", self.caddyfile)
        self.assertIn('respond "not found" 404', self.caddyfile)

    def test_overlay_publishes_only_https_frontend_not_internal_bridge(self):
        caddy = self.compose["services"]["caddy"]
        self.assertEqual(caddy["ports"], ["80:80/tcp", "443:443/tcp"])
        self.assertNotIn("8080:8080", str(self.compose))
        self.assertNotIn("27017:27017", str(self.compose))
        self.assertIn("wsc-authentication", caddy["depends_on"])

    def test_bootstrap_and_guide_keep_pretendo_pnids(self):
        self.assertIn("WSC_BRIDGE_DOMAIN", self.bootstrap)
        self.assertIn("60014/udp", self.bootstrap)
        self.assertIn("60015/udp", self.bootstrap)
        self.assertIn("Pretendo Network accounts", self.guide)
        self.assertIn("public Pretendo account route", self.guide)


if __name__ == "__main__":
    unittest.main()
