# HTTP to SSH driver stub
## Introduction
In order for ***Shifu*** to integrate with your driver. We have implemented a simple HTTP-to-SSH stub written in Go for developers to try out.

### Design
This HTTP to SSH stub is designed the following way:
- A SSH connection is made from the stub to the container itself, using the public key specified
- The SSH session is used as a reverse HTTP proxy which forwards to the localhost's specified HTTP port
- The stub will execute content in the HTTP request body directly in the SSH session
- The stub will proxy the result and execution status back to the requestor

### Functionality
#### Proxy HTTP body to SSH shell and execute
The main function for this stub is to take whatever passed to it in the HTTP body and issue the command with a specified timeout.

For example:

When using CURL to post a request to a given URL, the command looks like the following:

`curl -X POST -d "ping 8.8.8.8" http://example.com`

The request will then passes from the HTTP stub into the `shell` of the driver container:

`~ # ping 8.8.8.8`

And the result will look like the following from the HTTP client side (Note that the output is incomplete, this is due to the timeout environmental variable.):

```
PING 8.8.8.8 (8.8.8.8): 56 data bytes
64 bytes from 8.8.8.8: seq=0 ttl=36 time=47.227 ms
64 bytes from 8.8.8.8: seq=1 ttl=36 time=50.137 ms
64 bytes from 8.8.8.8: seq=3 ttl=36 time=47.619 ms
```

#### Check the `session.Run(cmd)` error and set the HTTP return status code
Currently it returns `200` if success and `400` for any error and timeout.

For errors, it will return both the `stdout` and `stderr` back inside the HTTP response body.

### Usage
We have written a sample Dockerfile [`driver_util/examples/simple-alpine/Dockerfile.sample`](/examples/driver_util/simple-alpine/Dockerfile.sample) which demonstrates how you can add the stub into an existing Alpine Docker image

The packaged Docker image takes the following environmental variables, so we need to configure them in the [yaml file](/examples/driver_util/simple-alpine/driver.yaml):
- `EDGEDEVICE_DRIVER_SSH_KEY_PATH`
  - The key path of SSH key on driver container which we used to connect to the driver container itself
- `EDGEDEVICE_DRIVER_HTTP_PORT` (Optional)
  - The HTTP server port of the driver container, default to `11112`
- `EDGEDEVICE_DRIVER_EXEC_TIMEOUT_SECOND` (Optional)
  - The timeout of an execution, this is achieved by appending `timeout <seconds>` in front of the command
- `EDGEDEVICE_DRIVER_SSH_USER` (Optional)
  - This is the user we used to SSH into the driver container, default to `root`
 