import subprocess
import sys

# Default value to write to the mock device
write_data = "88.8"

def run_command(command):
    try:
        result = subprocess.run(command, shell=True, check=True, capture_output=True, text=True)
        return result.stdout.strip()
    except subprocess.CalledProcessError as e:
        print(f"Command failed: {e}")
        return None

def get_pod_name():
    command = "kubectl get pods -n deviceshifu -l app=deviceshifu-lwm2m-deployment -o jsonpath='{.items[0].metadata.name}'"
    return run_command(command)

def curl_with_retry(url, method="GET", data=None):
    command = f"kubectl exec -n deviceshifu nginx -- curl --retry 5 --retry-delay 3 --retry-max-time 15 --connect-timeout 5 -X {method} {url}"
    if data:
        command += f" -d {data}"
    return run_command(command)

def main():
    # Get the pod name of deviceshifu
    pod_name = get_pod_name()

    if not pod_name:
        print("No deviceshifu-lwm2m pod found. Exiting...")
        sys.exit(1)

    # Use curl with retry options to retrieve information from the LwM2M server
    out = curl_with_retry("deviceshifu-lwm2m.deviceshifu.svc.cluster.local/float_value")

    # Check if the command succeeded
    if not out:
        print("Failed to retrieve value from the LwM2M server. Exiting...")
        sys.exit(1)

    # Remove any whitespace and newline characters
    out = out.replace("\r", "").replace("\n", "")
    print(f"Received value: {out}")

    # Check if the server response indicates an error
    if out == "Error on reading object":
        print("Device is unhealthy")
        run_command(f"kubectl logs -n deviceshifu {pod_name}")
        print("Timeout")
        sys.exit(1)

    # Use deviceshifu to write data to the mock device with retry settings
    curl_with_retry("deviceshifu-lwm2m.deviceshifu.svc.cluster.local/float_value", method="PUT", data=write_data)

    # Retrieve the value again after writing to verify if it was successful
    out = curl_with_retry("deviceshifu-lwm2m.deviceshifu.svc.cluster.local/float_value")
    
    # Check if the command succeeded
    if not out:
        print("Failed to verify the written value. Exiting...")
        sys.exit(1)

    out = out.replace("\r", "").replace("\n", "")

    # Check if the modification was successful
    if out == write_data:
        print("Modification successful")
        sys.exit(0)
    else:
        print(f"Modification failed, expected: {write_data}, got: {out}")
        sys.exit(1)

if __name__ == "__main__":
    main()
