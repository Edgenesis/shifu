#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Get the pod name of deviceshifu
pod_name=$(kubectl get pods -n deviceshifu -l app=deviceshifu-thermometer-deployment -o jsonpath='{.items[0].metadata.name}')

# Execute a curl command inside the nginx pod in the deviceshifu namespace to read the value from the thermometer device service
out=$(kubectl exec -n deviceshifu nginx -- curl --retry 5 --retry-delay 3 --max-time 15 --connect-timeout 5 thermometer.devices.svc.cluster.local:11111/read_value)

# Check if the kubectl exec command was successful
if [ $? -ne 0 ]; then
    echo "Failed to execute kubectl command"
    exit 1
fi

# Remove carriage return and newline characters from the output
out=$(echo "$out" | tr -d '\r\n')

# Print the received value
echo "Received value: $out"

# Check if the output indicates an error
if [[ $out == "Error on reading object" ]]; then
    echo "Device is unhealthy"
    # Ensure $pod_name is set before using it
    if [ -z "$pod_name" ]; then
        echo "\$pod_name is not set"
        exit 1
    fi
    # Print the logs of the pod
    kubectl logs -n deviceshifu $pod_name
    echo "Timeout"
    # Exit with status 1 to indicate failure
    exit 1
fi
