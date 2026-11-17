#!/usr/bin/env python3
"""Regression fixtures for the MQTT E2E shell checker (not live E2E)."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

SCRIPT = Path(__file__).with_name("checkoutput.sh")
DEFAULT_TIMESTAMP = "0001-01-01 00:00:00 +0000 UTC"
VALID = {"mqtt_message": "test2333", "mqtt_receive_timestamp": "2026-01-01 00:00:01 +0000 UTC"}


class CheckOutputTest(unittest.TestCase):
    def run_checker(self, responses, publish_status=0, curl_status=0):
        with tempfile.TemporaryDirectory(prefix="mqtt-check-fixture-") as directory:
            root = Path(directory)
            kubectl = root / "kubectl"
            kubectl.write_text("""#!/usr/bin/env python3
import json, os, pathlib, sys
root = pathlib.Path(os.environ['FIXTURE_DIR'])
args = sys.argv[1:]
with (root / 'calls.jsonl').open('a') as stream:
    stream.write(json.dumps(args) + '\\n')
if 'mosquitto_pub' in args:
    sys.exit(int(os.environ['PUBLISH_STATUS']))
if 'curl' in args:
    count_file = root / 'count'
    count = int(count_file.read_text()) if count_file.exists() else 0
    responses = json.loads(os.environ['RESPONSES'])
    print(responses[min(count, len(responses) - 1)])
    count_file.write_text(str(count + 1))
    sys.exit(int(os.environ['CURL_STATUS']))
sys.exit(127)
""")
            kubectl.chmod(0o755)
            sleep = root / "sleep"
            sleep.write_text("#!/bin/sh\nexit 0\n")
            sleep.chmod(0o755)
            env = dict(os.environ, PATH=f"{root}:{os.environ['PATH']}", FIXTURE_DIR=str(root), RESPONSES=json.dumps(responses), PUBLISH_STATUS=str(publish_status), CURL_STATUS=str(curl_status))
            result = subprocess.run(["bash", str(SCRIPT)], env=env, capture_output=True, text=True, timeout=15)
            calls = [json.loads(line) for line in (root / "calls.jsonl").read_text().splitlines()]
            return result, calls

    def test_exact_message_uses_configured_topic_and_instruction(self):
        result, calls = self.run_checker([json.dumps(VALID)])
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        publish = next(call for call in calls if "mosquitto_pub" in call)
        request = next(call for call in calls if "curl" in call)
        self.assertIn("deployment/mosquitto", publish)
        self.assertEqual(publish[publish.index("-t") + 1], "/test/test1")
        self.assertIn("http://deviceshifu-mqtt/get_topicmsg1", request)
        self.assertNotIn("-it", publish + request)

    def test_publish_failure_is_not_success(self):
        result, _ = self.run_checker([json.dumps(VALID)], publish_status=1)
        self.assertNotEqual(result.returncode, 0)

    def test_http_failure_is_not_success(self):
        result, _ = self.run_checker([json.dumps(VALID)], curl_status=22)
        self.assertNotEqual(result.returncode, 0)

    def test_error_page_is_not_success(self):
        result, _ = self.run_checker(["Error: Device instruction does not exist!"])
        self.assertNotEqual(result.returncode, 0)

    def test_malformed_json_is_not_success(self):
        result, _ = self.run_checker(['{"mqtt_message":'])
        self.assertNotEqual(result.returncode, 0)

    def test_wrong_message_is_not_success(self):
        result, _ = self.run_checker([json.dumps(dict(VALID, mqtt_message="other"))])
        self.assertNotEqual(result.returncode, 0)

    def test_missing_timestamp_is_not_success(self):
        result, _ = self.run_checker([json.dumps({"mqtt_message": "test2333"})])
        self.assertNotEqual(result.returncode, 0)

    def test_default_timestamp_is_not_success(self):
        result, _ = self.run_checker([json.dumps(dict(VALID, mqtt_receive_timestamp=DEFAULT_TIMESTAMP))])
        self.assertNotEqual(result.returncode, 0)

    def test_empty_response_is_not_success(self):
        result, _ = self.run_checker([""])
        self.assertNotEqual(result.returncode, 0)

    def test_waits_for_asynchronous_delivery(self):
        waiting = json.dumps({"mqtt_message": "", "mqtt_receive_timestamp": DEFAULT_TIMESTAMP})
        result, calls = self.run_checker([waiting, json.dumps(VALID)])
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertEqual(sum("curl" in call for call in calls), 2)


if __name__ == "__main__":
    unittest.main()
