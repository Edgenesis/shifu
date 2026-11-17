#!/usr/bin/env bash
set -euo pipefail

message='test2333'
# Publish using the already-ready broker pod, not a second broker with default config.
kubectl exec -n devices deployment/mosquitto -- mosquitto_pub \
    -h mosquitto-service -p 18830 -t /test/test1 -q 1 -m "$message"

# MQTT delivery is asynchronous; only the configured instruction's exact payload passes.
for attempt in {1..5}; do
    response=$(kubectl exec -n deviceshifu nginx -- curl \
        --fail --silent --show-error --connect-timeout 5 --max-time 10 \
        http://deviceshifu-mqtt/get_topicmsg1)
    printf '%s\n' "$response"
    if python3 -c '
import json
import sys
try:
    body = json.load(sys.stdin)
except ValueError:
    sys.exit(1)
if not isinstance(body, dict):
    sys.exit(1)
timestamp = body.get("mqtt_receive_timestamp")
valid = (body.get("mqtt_message") == sys.argv[1]
         and isinstance(timestamp, str) and bool(timestamp)
         and not timestamp.startswith("0001-01-01"))
sys.exit(0 if valid else 1)
' "$message" <<< "$response"; then
        printf 'MQTT message verified through get_topicmsg1\n'
        exit 0
    fi
    sleep 1
done
printf 'MQTT response did not contain the published message and receive timestamp\n' >&2
exit 1
