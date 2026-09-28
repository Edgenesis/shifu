#!/usr/bin/env python3
"""Retag in an isolated copy; publish only after generation and validation.

No Git commands or commits. Generation/download errors never touch the source
worktree. Publication errors roll back already replaced files. Like other
multi-file filesystem transactions, SIGKILL/power loss during publication is
not recoverable automatically; do not run concurrent release writers.
"""
import argparse
import os
from pathlib import Path
import shutil
import signal
import stat
import subprocess
import sys
import tempfile

sys.dont_write_bytecode = True
from validate_version import ROOT, check_target, references, validate

IGNORED = {".git", "__pycache__", ".venv", "node_modules"}


def snapshot(root):
    result = {}
    for directory, dirs, files in os.walk(root):
        dirs[:] = sorted(d for d in dirs if d not in IGNORED and d != "bin")
        for name in files:
            path = Path(directory) / name
            if name in IGNORED:
                continue
            if path.is_symlink():
                raise ValueError(f"refusing symlink in release tree: {path}")
            result[path.relative_to(root)] = (path.read_bytes(), stat.S_IMODE(path.stat().st_mode))
    return result


def rewrite(root, version):
    refs, errors = references(root)
    if errors:
        raise ValueError("\n".join(errors))
    by_path = {}
    for ref in refs:
        by_path.setdefault(ref.path, []).append(ref)
    for path, items in by_path.items():
        text = path.read_text()
        for ref in sorted(items, key=lambda item: item.start, reverse=True):
            text = text[:ref.start] + ref.target(version) + text[ref.end:]
        path.write_text(text)
    (root / "version.txt").write_text(version + "\n")
    return {p.relative_to(root) for p in by_path} | {Path("version.txt")}


def replace_file(path, state):
    if state is None:
        path.unlink(missing_ok=True)
        return
    data, mode = state
    path.parent.mkdir(parents=True, exist_ok=True)
    fd, temporary = tempfile.mkstemp(prefix=".retag-", dir=path.parent)
    try:
        with os.fdopen(fd, "wb") as stream:
            stream.write(data)
            stream.flush()
            os.fchmod(stream.fileno(), mode)
        os.replace(temporary, path)
    finally:
        if os.path.exists(temporary):
            os.unlink(temporary)


def publish(root, before, after, expected):
    changed = sorted(p for p in before.keys() | after.keys() if before.get(p) != after.get(p))
    current = snapshot(root)
    for path in changed:
        if current.get(path) != before.get(path):
            raise ValueError(f"concurrent edit detected: {path}; nothing published")
    applied = []
    try:
        for path in changed:
            applied.append(path)
            replace_file(root / path, after.get(path))
        errors = validate(root, expected)
        if errors:
            raise ValueError("\n".join(errors))
    except BaseException as failure:
        rollback_errors = []
        for path in reversed(applied):
            try:
                replace_file(root / path, before.get(path))
            except OSError as error:
                rollback_errors.append(f"{path}: {error}")
        if rollback_errors:
            raise RuntimeError("rollback incomplete; restore these files before releasing: "
                               + "; ".join(rollback_errors)) from failure
        raise
    return len(changed)


def retag(root, version, make="make"):
    check_target(version)
    before = snapshot(root)
    # Temp tree has no .git pointer back to the real worktree. Do not hardlink:
    # generators rewrite files in place and would corrupt the originals.
    with tempfile.TemporaryDirectory(prefix="shifu-retag-") as temporary:
        stage = Path(temporary) / "tree"
        shutil.copytree(root, stage, ignore=shutil.ignore_patterns(*IGNORED))
        allowed = rewrite(stage, version)
        env = dict(os.environ)
        # Inherited -j can race the install concatenation against generation;
        # inherited -n/-i/-k can turn failures into incomplete output.
        for name in ("MAKEFLAGS", "MFLAGS", "MAKEOVERRIDES"):
            env.pop(name, None)
        for target in ("generate-controller-yaml", "generate-install-yaml"):
            subprocess.run([make, "-j1", target, f"IMG=edgehub/shifu-controller:{version}"],
                           cwd=stage / "pkg/k8s/crd", env=env, check=True)
        errors = validate(stage, version)
        if errors:
            raise ValueError("\n".join(errors))
        after = snapshot(stage)
        for path in before.keys() | after.keys():
            if before.get(path) == after.get(path) or path in allowed:
                continue
            # controller-gen produces CRDs/RBAC/deepcopy; go fmt can format
            # controller sources. Never publish binaries, caches or other edits.
            if not (path.as_posix().startswith("pkg/k8s/") and path.suffix in (".go", ".yaml", ".yml")):
                raise ValueError(f"generator unexpectedly modified {path}; nothing published")
        return publish(root, before, after, version)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--version", required=True)
    parser.add_argument("--root", type=Path, default=ROOT)
    parser.add_argument("--make", default="make")
    args = parser.parse_args()

    def interrupted(signum, frame):
        raise InterruptedError(f"interrupted by signal {signum}")

    signal.signal(signal.SIGTERM, interrupted)
    try:
        count = retag(args.root.resolve(), args.version, args.make)
    except (OSError, ValueError, RuntimeError, subprocess.CalledProcessError, KeyboardInterrupt) as error:
        print(f"Retag failed: {error}", file=sys.stderr)
        return 1
    print(f"Retagged and validated {args.version}: {count} files updated")
    return 0


if __name__ == "__main__":
    sys.exit(main())
