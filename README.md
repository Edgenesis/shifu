- [Shifu](#shifu)
  - [What is Shifu?](#what-is-shifu)
  - [Why use Shifu?](#why-use-shifu)
  - [How to use Shifu?](#how-to-use-shifu)
  - [Quick Start Guide with Demo](#quick-start-guide-with-demo)
- [Shifu's vision](#shifus-vision)
  - [Make developers and operators happy again](#make-developers-and-operators-happy-again)
  - [Software Defined World (SDW)](#software-defined-world-sdw)

# Shifu

## What is Shifu?

Shifu is a framework designed to abstract out the complexity of interacting with IoT devices. Shifu aims to achieve TRUE plug'n'play IoT device management.

## Why use Shifu?

Shifu let you manage and control your IoT devices extremely easily. Whenever you connect your device, Shifu will recognize it and spawn an augmented digital twin called ***deviceShifu*** for it. deviceShifu provides you with a high-level abstraction to interact with your device. By implementing the interface of the deviceShifu, your IoT device can achieve everything its designed for, and much more! For example, your device's state can be rolled back with a single line of command. (If physically permitted, of course.) Furthermore, Shifu will secure your entire IoT system from ground up. 

## How to use Shifu?

Currently, Shifu runs on [Kubernetes](k8s.io). We will provide more deployment methods including standalone deployment in the future.

## Quick Start Guide with Demo


###	Prerequisite:
The following services and tools are needed to be installed before running shifu demo:
1. [Docker](https://docs.docker.com/get-docker/)
2. [Kubernetes](https://kubernetes.io/docs/setup/)
3. [kubebuilder](https://book.kubebuilder.io/quick-start.html)
4. [kind (kubernetes in docker)](https://github.com/kubernetes-sigs/kind)


### Edge device controller setup:
We have prepared several useful scripts for faster hands-on under `test/scripts`. The simplest way of getting started is this:
1. **start the cluster with CRD:**
	
	Create a local kubernetes cluster:

	```
	kind create cluster
	```
	
	Expected output should be similar to:
	```
	Creating cluster "kind" ...
	✓ Ensuring node image (kindest/node:v1.21.1) 🖼 
	✓ Preparing nodes 📦  
    ✓ Writing configuration 📜 
	✓ Starting control-plane 🕹️ 
	✓ Installing CNI 🔌 
    ✓ Installing StorageClass 💾 
	Set kubectl context to "kind-kind"
	You can now use your cluster with:

	kubectl cluster-info --context kind-kind

    Thanks for using kind! 😊
   ```
    
    setup CRD for edge device controller:
    ```
    ./test/scripts/deviceshifu-setup.sh apply
    ```
	After the that step, try 
	```
	% kubectl get pods --all-namespaces
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
2. **start demo devices:**
    In directory deviceshifu/examples/demo_device, we have 4 demo devices: `agv`, `robotarm`, `tecan`,`thermometer`.
Currently, those demo devices only have one instruction available: `get_status` - which will return a random status of the device such like Idle, Busy, Running, Error etc.

    Start the demo device pods:
    ```
    ./test/scripts/deviceshifu-demo.sh apply edgedevice-thermometer
    ./test/scripts/deviceshifu-demo.sh apply edgedevice-agv
    ./test/scripts/deviceshifu-demo.sh apply edgedevice-tecan
    ./test/scripts/deviceshifu-demo.sh apply edgedevice-robot-arm
    ```
    We should have the following pods running in devices namespace:
    ```
    % kubectl get pods --namespace devices
    NAME                           READY   STATUS    RESTARTS   AGE
    agv-5944698b79-qxdmk           1/1     Running   0          86s
    robotarm-5478f86fc8-s5kmg      1/1     Running   0          85s
    tecan-6859f67bc5-htxpp         1/1     Running   0          86s
    thermometer-6d6d8f759f-4hd6l   1/1     Running   0          28m
    ```
    With a `kubectl proxy` we would be able to see the pods information, which is useful for verifying the active pod configurations.
    
    Additionally, as the pod IPs are not routable, we can run a simple nginx server within the cluster under `devices` namespace, and `kubectl exec` into the nginx pod. 
    After that, we can call get_status on those mock devices and see the return value:
    ```
    / # curl http://thermometer:11111/get_status
    Running
    
    / # curl http://agv:11111/get_status
    Busy
    
    / # curl http://tecan:11111/get_status
    Busy
    
    / # curl http://robotarm:11111/get_status
    Idle
    ```

# Shifu's vision

## Make developers and operators happy again

Developers and operators should 100% focus on using their creativity to make huge impacts, not fixing infrastructures and re-inventing the wheels day in and day out. As a developer and an operator, Shifu knows your pain!!! Shifu wants to take out all the underlying mess, and make us developers and operators happy again!

## Software Defined World (SDW)

If every IoT device has a Shifu with it, we can leverage software to manage our surroundings, and make the whole world software defined. In a SDW, everything is intelligent, and the world around you will automatically change itself to serve your needs. After all, all the technology we built are designed to serve human beings. 
