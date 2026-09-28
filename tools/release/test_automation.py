#!/usr/bin/env python3
"""Offline shell regression fixtures; no GitHub requests or mutations."""
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

HELPER = Path(__file__).with_name("automation.sh")


class MergeGateTests(unittest.TestCase):
    def run_gate(self, mode):
        with tempfile.TemporaryDirectory() as directory:
            log = Path(directory) / "mutations"
            script = r'''set -euo pipefail
source "$HELPER"
gh() {
  case "$1 $2" in
    "pr view")
      case "$*" in
        *headRefOid*)
          if [[ "$MODE" == changed_sha && -e "$MUTATIONS.checks" ]]; then printf '%040d\n' 2; else printf '%040d\n' 1; fi ;;
        *mergedAt*) return 1 ;;
        *reviewDecision*) echo NONE ;;
        *)
          [[ "$MODE" != refresh_error ]] || return 1
          if [[ "$MODE" == unknown ]]; then echo UNKNOWN; else echo CLEAN; fi ;;
      esac ;;
    "pr checks")
      if [[ -e "$MUTATIONS.checks" ]]; then
        case "$MODE" in
          admin_red) echo '[{"bucket":"fail","name":"ci"}]'; return 0 ;;
          admin_unknown) echo '[{"bucket":"mystery","name":"ci"}]'; return 0 ;;
        esac
      fi
      touch "$MUTATIONS.checks"
      case "$MODE" in
        red) echo '[{"bucket":"fail","name":"ci","state":"FAILURE"}]'; return 1 ;;
        red_json) echo '[{"bucket":"fail","name":"ci","state":"FAILURE"}]' ;;
        unknown_checks) echo '[{"bucket":"mystery","name":"ci"}]' ;;
        unreadable) echo 'not json'; return 1 ;;
        *) echo '[{"bucket":"pass","name":"ci"}]' ;;
      esac ;;
    "pr merge")
      echo "$*" >> "$MUTATIONS"
      if [[ "$MODE" == merge_error ]]; then echo 'merge failed' >&2; return 1; fi
      if [[ "$MODE" == admin_* || "$MODE" == auto_wait_failure ]]; then
        [[ "$*" != *--auto* ]] || return 0
        echo 'base branch policy prohibits the merge' >&2; return 1
      fi ;;
    *) echo "$*" >> "$MUTATIONS" ;;
  esac
}
# Errexit is deliberately ineffective in this captured conditional caller.
if output="$(release_merge_pr_when_ready fixture/repo 1 2 1)"; then
  exit 0
else
  exit "$?"
fi
'''
            env = dict(os.environ, HELPER=str(HELPER), MODE=mode, MUTATIONS=str(log))
            result = subprocess.run(["bash", "-c", script], env=env, text=True, capture_output=True, timeout=10)
            return result, log.read_text() if log.exists() else ""

    def test_no_merge_after_red_unknown_or_unreadable_checks(self):
        for mode in ("red", "red_json", "unknown", "unknown_checks", "unreadable", "refresh_error", "changed_sha"):
            with self.subTest(mode=mode):
                result, mutations = self.run_gate(mode)
                self.assertNotEqual(result.returncode, 0, result.stdout + result.stderr)
                self.assertEqual(mutations, "")

    def test_merge_failure_propagates_in_substitution(self):
        result, mutations = self.run_gate("merge_error")
        self.assertNotEqual(result.returncode, 0)
        self.assertNotIn("--admin", mutations)

    def test_admin_merge_rechecks_fail_closed(self):
        for mode in ("admin_red", "admin_unknown"):
            with self.subTest(mode=mode):
                result, mutations = self.run_gate(mode)
                self.assertNotEqual(result.returncode, 0, result.stderr)
                self.assertEqual(len(mutations.splitlines()), 1)
                self.assertNotIn("--admin", mutations)
                self.assertNotIn("--auto", mutations)

    def test_auto_merge_wait_failure_propagates(self):
        result, mutations = self.run_gate("auto_wait_failure")
        self.assertNotEqual(result.returncode, 0, result.stderr)
        self.assertIn("--auto", mutations)

    def test_green_checked_sha_merges(self):
        result, mutations = self.run_gate("green")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("--match-head-commit", mutations)
        self.assertNotIn("--admin", mutations)


class WorkflowOwnershipTests(unittest.TestCase):
    def test_legacy_tag_workflows_are_validation_only(self):
        directory = HELPER.parents[2] / ".github/workflows"
        for name in ("release-rc-tag.yml", "release-official-tag.yml"):
            with self.subTest(workflow=name):
                text = (directory / name).read_text()
                self.assertIn("contents: read", text)
                self.assertIn("persist-credentials: false", text)
                self.assertIn("ref: ${{ github.event.pull_request.merge_commit_sha }}", text)
                self.assertIn("validate_version.py --expected", text)
                for forbidden in ("contents: write", "gh release create", "git tag ", "git push", "dispatches", "create-github-app-token", "GH_TOKEN"):
                    self.assertNotIn(forbidden, text)

    def test_candidate_ci_runs_regressions_and_version_validation(self):
        text = (HELPER.parents[2] / ".github/workflows/release-validation.yml").read_text()
        self.assertIn("pull_request:", text)
        self.assertIn("unittest discover -s tools/release", text)
        self.assertIn('validate_version.py --expected "$(cat version.txt)"', text)


if __name__ == "__main__":
    unittest.main()
