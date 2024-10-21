#!/bin/bash

writeData=88.8

# Retrieve LwM2M server information with multiple retries
for i in {1..15}; do
    # Check deviceshifu status
    out=$(kubectl exec -it -n deviceshifu nginx -- curl deviceshifu-lwm2m.deviceshifu.svc.cluster.local/float_value --connect-timeout 5)
    
    # Remove any whitespace and newline characters
    out=$(echo "$out" | tr -d '\r\n')

    # Output the status
    echo "Received value: $out"
    
    # Check if the server response is empty
    if [[ $out == "Error on reading object" ]]; then
        echo "Device is unhealthy"
    else
        break
    fi

    # If after 5 attempts there is still no success, timeout and exit
    if [[ $i -eq 5 ]]; then
        echo "timeout"
        exit 1
    fi

    # Wait for some time before retrying
    sleep 1
done

# Use deviceshifu to write data to the mock device
kubectl exec -it -n deviceshifu nginx -- curl -X PUT deviceshifu-lwm2m.deviceshifu.svc.cluster.local/float_value -d $writeData --connect-timeout 5

# Retrieve the value after writing to verify if it was successful
out=$(kubectl exec -it -n deviceshifu nginx -- curl deviceshifu-lwm2m.deviceshifu.svc.cluster.local/float_value --connect-timeout 5)

# Remove any whitespace and newline characters
out=$(echo "$out" | tr -d '\r\n')

# Check if the modification was successful
if awk "BEGIN {exit !($out == $writeData)}"; then
    echo "modify success"
    exit 0
else
    echo "modify failed, expected: $writeData, got: $out"
fi
