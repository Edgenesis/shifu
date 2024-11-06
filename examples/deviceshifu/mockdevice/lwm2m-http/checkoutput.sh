#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Get the pod name of deviceshifu
pod_name=$(kubectl get pods -n deviceshifu -l app=deviceshifu-lwm2m-deployment -o jsonpath='{.items[0].metadata.name}')

get_value() {
    kubectl exec -n deviceshifu nginx -- curl --connect-timeout 5 http://deviceshifu-lwm2m-service.deviceshifu.svc.cluster.local/read_value
}

# Attempt to get the value with retries
for i in {1..15}; do
    out=$(get_value)
    
    # Remove any whitespace and newline characters
    out=$(echo "$out" | tr -d '\r\n')
    
    # Output the status
    echo "Received value: $out"
    
    # Check if the server response indicates an error
    if [[ -n "$out" && $out != "Error on reading object" ]]; then
        break
    fi
    
    echo "Device is unhealthy. Attempting to reconnect... ($i/15)"
    sleep 3
done

if [[ -z "$out" || $out == "Error on reading object" ]]; then
    echo "Device is still unhealthy after 15 attempts. Exiting..."
    kubectl logs -n deviceshifu $pod_name
    exit 1
fi
