#!/usr/bin/env python3
"""Check release-owned references without dependencies or network access.

Keep the image allowlist and path scope shared with retag.py. Documentation,
ConfigMap driverImage examples, third-party/custom images and the historical
v0.0.1 mockdevice fixture are deliberately not release outputs.
"""
import argparse
from dataclasses import dataclass
from pathlib import Path
import re
import sys

ROOT = Path(__file__).resolve().parents[2]
IMAGES = frozenset("edgehub/" + name for name in (
    "shifu-controller", "telemetryservice", "deviceshifu-http-http",
    "deviceshifu-http-mqtt", "deviceshifu-http-socket", "deviceshifu-http-opcua",
    "deviceshifu-http-lwm2m", "deviceshifu-tcp-tcp", "gateway-lwm2m",
    "mockdevice-thermometer", "mockdevice-robot-arm", "mockdevice-plate-reader",
    "mockdevice-agv", "mockdevice-plc", "mockdevice-socket", "mockdevice-opcua",
    "deviceshifu-http-http-python", "humidity-detector", "mockserver",
))
PINNED = {"examples/deviceshifu/mockdevice/test-edgedevice-mockdevice-deployment.yaml"}
CONTROLLER = "pkg/k8s/controllers/telemetryservice_controller.go"
AIO = "test/scripts/deviceshifu-demo-aio.sh"
KUSTOMIZATION = "pkg/k8s/crd/config/manager/kustomization.yaml"
REQUIRED_IMAGES = {
    "pkg/k8s/crd/install/config_default.yaml": "edgehub/shifu-controller",
    "pkg/k8s/crd/install/shifu_install.yml": "edgehub/shifu-controller",
    "pkg/telemetryservice/install/telemetryservice_install.yaml": "edgehub/telemetryservice",
}
IMAGE_FIELD = re.compile(r'''(?m)^\s*(?:-\s*)?image:\s*["']?(?P<value>[^\s"'#]+)''')
GO_IMAGE = re.compile(r'(?m)^const IMAGE\s*=\s*"(?P<value>[^"]+)"')
AIO_VERSION = re.compile(r'''(?m)^SHIFU_IMG_VERSION=["']?(?P<value>[^\s"'#]+)''')
MANAGER_IMAGE = re.compile(
    r'(?m)^\s*newName:\s*edgehub/shifu-controller\s*\n\s*newTag:\s*[\"\']?(?P<value>[^\s\"\']+)')


@dataclass(frozen=True)
class Reference:
    path: Path
    start: int
    end: int
    value: str
    repository: str = ""

    def target(self, version):
        return f"{self.repository}:{version}" if self.repository else version


def check_target(version):
    if not re.fullmatch(r"[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}", version):
        raise ValueError("expected a nonempty Docker image tag (e.g. v0.106.1, v0.106.1-rc1 or nightly)")


def manifest_paths(root):
    paths = set()
    for directory in ("examples", "pkg/k8s/crd/install", "pkg/telemetryservice/install"):
        for path in (root / directory).rglob("*"):
            if path.suffix not in (".yaml", ".yml"):
                continue
            rel = path.relative_to(root).as_posix()
            if rel in PINNED:
                continue
            if directory != "examples" or "deployment" in path.name or "install" in path.name:
                paths.add(path)
    return sorted(paths)


def references(root):
    refs, errors = [], []
    for path in manifest_paths(root):
        text = path.read_text()
        for match in IMAGE_FIELD.finditer(text):
            value = match["value"]
            repository = re.split(r"[:@]", value, maxsplit=1)[0]
            if repository in IMAGES:
                refs.append(Reference(path, *match.span("value"), value, repository))
    for rel, repository in REQUIRED_IMAGES.items():
        if not any(r.path == root / rel and r.repository == repository for r in refs):
            errors.append(f"{rel}: missing required {repository} image")
    for rel, pattern, repository in (
        (CONTROLLER, GO_IMAGE, "edgehub/telemetryservice"),
        (AIO, AIO_VERSION, ""),
        (KUSTOMIZATION, MANAGER_IMAGE, ""),
    ):
        path = root / rel
        matches = list(pattern.finditer(path.read_text())) if path.is_file() else []
        if len(matches) != 1:
            errors.append(f"{rel}: expected exactly one release version reference, found {len(matches)}")
        else:
            match = matches[0]
            refs.append(Reference(path, *match.span("value"), match["value"], repository))
    return refs, errors


def validate(root, expected):
    check_target(expected)
    refs, errors = references(root)
    path = root / "version.txt"
    actual = path.read_text().strip() if path.is_file() else "<missing>"
    if actual != expected:
        errors.append(f"version.txt: expected {expected}, found {actual}")
    for ref in refs:
        if ref.value != ref.target(expected):
            errors.append(f"{ref.path.relative_to(root)}: expected {ref.target(expected)}, found {ref.value}")
    return errors


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--expected", required=True)
    parser.add_argument("--root", type=Path, default=ROOT, help="repository root (defaults to this script's repository)")
    args = parser.parse_args()
    try:
        errors = validate(args.root.resolve(), args.expected)
    except (OSError, ValueError) as error:
        errors = [str(error)]
    if errors:
        print("Release version validation failed:\n" + "\n".join(errors), file=sys.stderr)
        return 1
    print(f"Release versions are consistent: {args.expected}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
