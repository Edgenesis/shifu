# Quickest Start
We have created an all-in-one docker image containing everything needed for the demo to run. 
As long as you have [Docker](https://docs.docker.com/get-docker/) installed, you can immediately start your shifu experience.

1. **Pull and run the docker container:**

    ```
    docker run --network host -it -v /var/run/docker.sock:/var/run/docker.sock edgehub/demo-image-alpine:v0.0.1
    ```

2. **Create kubernetes cluster and start shifu services:**
    
    The script below will create a kubernetes cluster with pre-defined CustomResourceDefinition, as well as starting the minimal required shifu services.
    ```
    ./test/scripts/deviceshifu-setup.sh apply
    ```
    After the that step, try 
    ```
    kubectl get pods --all-namespaces
    ```

    and we should have the following pods running like this:
    ```
    NAMESPACE            NAME                                         READY   STATUS    RESTARTS   AGE
    crd-system           crd-controller-manager-7bc78896b9-cpk7d      2/2     Running   0          11m
    kube-system          coredns-558bd4d5db-khlqs                     1/1     Running   0          13m
    kube-system          coredns-558bd4d5db-w4tvl                     1/1     Running   0          13m
    kube-system          etcd-kind-control-plane                      1/1     Running   0          13m
    kube-system          kindnet-75547                                1/1     Running   0          13m
    kube-system          kube-apiserver-kind-control-plane            1/1     Running   0          13m
    kube-system          kube-controller-manager-kind-control-plane   1/1     Running   0          13m
    kube-system          kube-proxy-g5kbl                             1/1     Running   0          13m
    kube-system          kube-scheduler-kind-control-plane            1/1     Running   0          13m
    local-path-storage   local-path-provisioner-547f784dff-wspb2      1/1     Running   0          13m
    ```
    we can check the logs to verify the running status:
    ```
    kubectl --namespace crd-system logs crd-controller-manager-7bc78896b9-cpk7d -c manager
    ```

3. **Start demo deviceShifu (digital twins):**
    
    In directory examples/deviceshifu/demo_device, we have 4 mock devices for shifu to create a ***deviceShifu*** (digital twin) for each; all devices have a `get_status` instruction which returns the current status of the device, such like Busy, Error, Idle, etc.
    Besides `get_status`, each mock device has its own instruction:
    * **thermometer**: a thermometer which reports a temperature, it has an instruction `read_value` which returns the temperature value read from the thermometer
    * **agv**: an automated guided vehicle, it has two instructions, it has an instruction `get_position` which returns the position of the vehicle in the x-y coordinate
    * **robotarm**: a robot arm used in a lab, it has an insturuction of `get_coordinate` which returns the position of the robot arm in the x-y-z coordinate
    * **plate-reader**: a plate with a matrix of blocks where each block contains some liquid, it has an insturuction of `get_measurement` which returns a matrix of the measurement value of the liquid in each block

    Start the deviceShifu for all 4 devices:
    ```
    kubectl apply -f examples/deviceshifu/demo_device/edgedevice-thermometer
    kubectl apply -f examples/deviceshifu/demo_device/edgedevice-agv
    kubectl apply -f examples/deviceshifu/demo_device/edgedevice-robot-arm
    kubectl apply -f examples/deviceshifu/demo_device/edgedevice-plate-reader
    ```
    We should have the following pods running in devices namespace:
    ```
    kubectl get pods --namespace devices
    ```
    which should give us an output like this:
    ```
    NAME                           READY   STATUS    RESTARTS   AGE
    agv-5944698b79-qxdmk           1/1     Running   0          86s
    robotarm-5478f86fc8-s5kmg      1/1     Running   0          85s
    thermometer-6d6d8f759f-4hd6l   1/1     Running   0          28m
    plate-reader-6859f67bc5-htxpp         1/1     Running   0          86s
    ```
    We can view the info of a deviceShifu with `kubectl describe pods`. For example,
    ```
    kubectl describe pods edgedevice-thermometer-deployment-b648d5c6c-rf88p
    ```
4. **Start a nginx server and communicate with each deviceShifu:**
    
    As the pod IPs are not routable, we can run a simple nginx server, and `kubectl exec` into the nginx pod. 
    A nginx is provided within the container, so we can just start the pod:
    ```
    kubectl run nginx --image=nginx:1.21 -n deviceshifu
    ```
    Then we are able to get into the shell of the nginx server:
    ```
    kubectl -n deviceshifu exec -it nginx -- bash
    ```
    After that, we can call each deviceShifu's instructions on those mock devices and see the return value.
    Each deviceShifu has its instructions defined in the configmap file.
    ```

    / # curl http://deviceshifu-thermometer/get_status
    Busy
    / # curl http://deviceshifu-thermometer/read_value
    27
    / # curl http://deviceshifu-agv/get_status
    Busy
    / # curl http://deviceshifu-agv/get_position
    xpos: 54, ypos: 34
    / # curl http://deviceshifu-robotarm/get_status
    Busy
    / # curl http://deviceshifu-robotarm/get_coordinate
    xpos: 55, ypos: 140, zpos: 135
    / # curl http://deviceshifu-plate-reader/get_status
    Idle
    / # curl http://deviceshifu-plate-reader/get_measurement
    0.75 0.50 1.34 0.95 2.79 2.66 2.68 0.59 0.97 0.93 0.70 0.62 
    0.61 1.47 1.68 1.65 1.05 1.59 0.78 2.92 1.22 1.12 2.86 0.29 
    2.15 2.45 1.99 0.36 1.47 0.18 2.47 0.61 2.43 1.53 0.14 2.41 
    2.80 2.49 0.63 2.61 1.09 1.46 0.22 1.99 1.46 2.30 0.51 0.41 
    1.24 0.78 0.34 2.83 2.76 1.89 2.64 1.79 1.24 1.68 2.84 2.92 
    2.09 2.38 0.02 0.47 0.38 1.62 2.65 0.58 2.17 2.70 0.97 2.18 
    1.47 0.66 0.61 0.10 2.91 1.61 0.30 2.21 0.46 1.74 1.62 1.01 
    1.28 2.27 1.04 0.44 2.47 1.83 0.59 2.09 1.30 2.24 2.87 2.78 
    ```
