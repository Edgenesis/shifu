"""Offline release regression fixtures; no Docker, Go downloads or Git writes."""
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest
from unittest import mock

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(Path(__file__).resolve().parent))
import retag
from validate_version import references


class RetagTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name) / "repo"
        shutil.copytree(ROOT, self.root, ignore=shutil.ignore_patterns(
            ".git", "bin", "__pycache__", ".venv"))
        # Recreate the bad stable state explicitly so tests survive a retag of
        # the repository containing this suite.
        retag.rewrite(self.root, "v0.106.0")
        for rel in ("examples/httpDeviceShifu/deployment/http-deviceshifu-deployment.yaml",
                    "pkg/k8s/controllers/telemetryservice_controller.go"):
            path = self.root / rel
            path.write_text(path.read_text().replace(":v0.106.0", ":v0.106.0-rc1"))
        self.makefile = self.root / "pkg/k8s/crd/Makefile"
        # Minimal, explicitly synthetic generator with the real target interface.
        self.makefile.write_text('generate-controller-yaml:\n\t@true\n'
                                 'generate-install-yaml:\n\t@true\n')

    def snapshot(self):
        return {str(p.relative_to(self.root)): (p.read_bytes(), p.stat().st_mode)
                for p in self.root.rglob("*") if p.is_file()
                and "__pycache__" not in p.parts}

    def run_tag(self, version="v0.106.1", **kwargs):
        env = dict(os.environ, PYTHONDONTWRITEBYTECODE="1")
        return subprocess.run(["make", "tag", "VERSION=" + version],
                              cwd=self.root, env=env, text=True,
                              stdout=subprocess.PIPE, stderr=subprocess.STDOUT, **kwargs)

    def validate(self, version):
        return subprocess.run([sys.executable, "tools/release/validate_version.py",
                               "--expected", version], cwd=self.root,
                              text=True, capture_output=True)

    def test_existing_bad_stable_is_rejected(self):
        result = self.validate("v0.106.0")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("v0.106.0-rc1", result.stderr)
        self.assertIn("telemetryservice_controller.go", result.stderr)

    def test_rc_to_stable_and_idempotence(self):
        for version in ("v0.106.1-rc1", "v0.106.1"):
            result = self.run_tag(version)
            self.assertEqual(result.returncode, 0, result.stdout)
            self.assertEqual(self.validate(version).returncode, 0)
        before = self.snapshot()
        result = self.run_tag()
        self.assertEqual(result.returncode, 0, result.stdout)
        self.assertEqual(self.snapshot(), before)

    def test_mixed_tags_repaired_and_unrelated_images_preserved(self):
        target = self.root / "examples/mixed-deployment.yaml"
        unrelated = ('image: customer/deviceshifu-http-http:nightly\n'
                     'image: edgehub/camera-python:v0.0.1\n'
                     'image: nginx:1.21\n'
                     '# image: edgehub/deviceshifu-http-http:comment-only\n')
        target.write_text(unrelated + 'image: edgehub/deviceshifu-http-http:nightly\n'
                          'image: "edgehub/mockdevice-plc:v0.106.0-rc1"\n'
                          'image: edgehub/gateway-lwm2m:v0.105.0')
        pinned = self.root / "examples/deviceshifu/mockdevice/test-edgedevice-mockdevice-deployment.yaml"
        pinned_before = pinned.read_bytes()
        aio = self.root / "test/scripts/deviceshifu-demo-aio.sh"
        aio.write_text(aio.read_text() + '\nUNRELATED_VERSION=v0.106.0\n')
        result = self.run_tag()
        self.assertEqual(result.returncode, 0, result.stdout)
        self.assertTrue(target.read_text().startswith(unrelated))
        self.assertEqual(target.read_text().count(":v0.106.1"), 3)
        self.assertEqual(pinned.read_bytes(), pinned_before)
        self.assertIn("UNRELATED_VERSION=v0.106.0", aio.read_text())
        self.assertEqual(self.validate("v0.106.1").returncode, 0)

    def test_mid_generation_network_failure_changes_nothing(self):
        self.makefile.write_text('generate-controller-yaml:\n'
                                '\tprintf partial > install/config_default.yaml\n'
                                '\tprintf "go: download: unexpected EOF\\n" >&2\n'
                                '\texit 1\n'
                                'generate-install-yaml:\n\t@true\n')
        before = self.snapshot()
        result = self.run_tag()
        self.assertNotEqual(result.returncode, 0, result.stdout)
        self.assertIn("unexpected EOF", result.stdout)
        self.assertEqual(self.snapshot(), before)

    def test_failed_validation_changes_nothing(self):
        self.makefile.write_text('generate-controller-yaml:\n'
                                '\tprintf "image: edgehub/shifu-controller:wrong\\n" > install/config_default.yaml\n'
                                'generate-install-yaml:\n\t@true\n')
        before = self.snapshot()
        result = self.run_tag()
        self.assertNotEqual(result.returncode, 0, result.stdout)
        self.assertEqual(self.snapshot(), before)

    def test_publication_failure_rolls_back(self):
        before = retag.snapshot(self.root)
        after = dict(before)
        for path in (Path("version.txt"), Path("pkg/k8s/crd/install/config_default.yaml")):
            after[path] = (b"changed", before[path][1])
        original = retag.replace_file
        calls = 0

        def fail_second(path, state):
            nonlocal calls
            calls += 1
            if calls == 2:
                raise OSError("injected publication failure")
            original(path, state)

        with mock.patch.object(retag, "replace_file", side_effect=fail_second):
            with self.assertRaisesRegex(OSError, "injected publication failure"):
                retag.publish(self.root, before, after, "v0.106.1")
        self.assertEqual(retag.snapshot(self.root), before)

    def test_every_version_surface_is_checked(self):
        retag.rewrite(self.root, "v0.106.1")
        refs, errors = references(self.root)
        self.assertEqual(errors, [])
        for ref in refs:
            with self.subTest(path=ref.path):
                original = ref.path.read_text()
                ref.path.write_text(original[:ref.start] + ref.target("wrong") + original[ref.end:])
                self.assertNotEqual(self.validate("v0.106.1").returncode, 0)
                ref.path.write_text(original)
        (self.root / "version.txt").write_text("nightly\n")
        self.assertNotEqual(self.validate("v0.106.1").returncode, 0)

    def test_nightly_target(self):
        result = self.run_tag("nightly")
        self.assertEqual(result.returncode, 0, result.stdout)
        self.assertEqual(self.validate("nightly").returncode, 0)

    def test_missing_output_rejected(self):
        (self.root / "pkg/k8s/crd/install/shifu_install.yml").unlink()
        self.assertNotEqual(self.validate("v0.106.0").returncode, 0)

    def test_make_dry_run_changes_nothing(self):
        before = self.snapshot()
        result = subprocess.run(["make", "-n", "tag", "VERSION=v0.106.1"],
                                cwd=self.root, capture_output=True, text=True)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertEqual(self.snapshot(), before)

    def test_missing_or_invalid_target_changes_nothing(self):
        before = self.snapshot()
        for version in ("", "v1.2.3/invalid", "v1.2.3\nnightly"):
            self.assertNotEqual(self.run_tag(version).returncode, 0)
            self.assertEqual(self.snapshot(), before)


if __name__ == "__main__":
    unittest.main()
