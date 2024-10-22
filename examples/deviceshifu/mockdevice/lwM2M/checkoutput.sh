#!/bin/bash

#Default value to write to the mock device
writeData=88.8

#Get the pod name of deviceshifu
pod_name=$(kubectl get pods -n deviceshifu -l app=deviceshifu-lwm2m-deployment -o jsonpath='{.items[0].metadata.name}')

if [ -z "$pod_name" ]; then
    echo "No deviceshifu-lwm2m pod found. Exiting..."
    exit 1
fi

# Retrieve LwM2M server information with multiple retries
for i in {1..30}; do
    # Check deviceshifu status
    out=$(kubectl exec -n deviceshifu nginx -- curl deviceshifu-lwm2m.deviceshifu.svc.cluster.local/float_value --connect-timeout 5)
    # Remove any whitespace and newline characters
    out=$(echo "$out" | tr -d '\r\n')

    # Output the status
    echo "Received value: $out"
    
    # Check if the server response is empty
    if [[ $out == "Error on reading object" ]]; then
        echo "Device is unhealthy"
        kubectl logs -n deviceshifu $pod_name
    else
        break
    fi

    # If after 5 attempts there is still no success, timeout and exit
    if [[ $i -eq 5 ]]; then
        echo "timeout"
        exit 1
    fi

    # Wait for some time before retrying
    sleep 3
done

# Use deviceshifu to write data to the mock device
kubectl exec -n deviceshifu nginx -- curl -X PUT deviceshifu-lwm2m.deviceshifu.svc.cluster.local/float_value -d $writeData --connect-timeout 5

# Retrieve the value after writing to verify if it was successful
out=$(kubectl exec -n deviceshifu nginx -- curl deviceshifu-lwm2m.deviceshifu.svc.cluster.local/float_value --connect-timeout 5)

# Remove any whitespace and newline characters
out=$(echo "$out" | tr -d '\r\n')

# Check if the modification was successful
if awk "BEGIN {exit !($out == $writeData)}"; then
    echo "modify success"
    exit 0
else
    echo "modify failed, expected: $writeData, got: $out"
fi
