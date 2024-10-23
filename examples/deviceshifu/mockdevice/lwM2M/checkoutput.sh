#!/bin/bash

# Default value to write to the mock device
writeData=88.8

# Get the pod name of deviceshifu
pod_name=$(kubectl get pods -n deviceshifu -l app=deviceshifu-lwm2m-deployment -o jsonpath='{.items[0].metadata.name}')

if [ -z "$pod_name" ]; then
    echo "No deviceshifu-lwm2m pod found. Exiting..."
    exit 1
fi

# Function to retrieve value from the LwM2M server
get_value() {
    kubectl exec -n deviceshifu nginx -- curl --connect-timeout 5 deviceshifu-lwm2m.deviceshifu.svc.cluster.local/float_value
}

# Attempt to get the value with retries
for i in {1..5}; do
    out=$(get_value)
    
    # Remove any whitespace and newline characters
    out=$(echo "$out" | tr -d '\r\n')
    
    # Output the status
    echo "Received value: $out"
    
    # Check if the server response indicates an error
    if [[ $out != "Error on reading object" ]]; then
        break
    fi
    
    echo "Device is unhealthy. Attempting to reconnect... ($i/5)"
    sleep 3
done

if [[ $out == "Error on reading object" ]]; then
    echo "Device is still unhealthy after 5 attempts. Exiting..."
    kubectl logs -n deviceshifu $pod_name
    exit 1
fi

# Use deviceshifu to write data to the mock device with retry settings
kubectl exec -n deviceshifu nginx -- curl --retry 5 --retry-delay 3 --max-time 15 --connect-timeout 5 -X PUT deviceshifu-lwm2m.deviceshifu.svc.cluster.local/float_value -d $writeData

# Retrieve the value again after writing to verify if it was successful
out=$(get_value)

# Remove any whitespace and newline characters
out=$(echo "$out" | tr -d '\r\n')

# Check if the modification was successful
if awk "BEGIN {exit !($out == $writeData)}"; then
    echo "Modification successful"
    exit 0
else
    echo "Modification failed, expected: $writeData, got: $out"
    exit 1
fi
