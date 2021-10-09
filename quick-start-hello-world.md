# Hello World Device
This section will show you how *shifu* works by creating a simple edge device and connect to *shifu* with its *deviceShifu* (digital twin).
This edge device will only do one thing: response a message upon the HTTP endpoint /hello.

1. ### Prepare the edge device docker image
   The expected edge device is a http server that response "Hello_world from device via shifu!"
   In your working directory, for example, create a `helloworld.go` with following content:
   ```
		package main
		import (
			"fmt"
			"net/http"
		)

		func process_hello(w http.ResponseWriter, req *http.Request) {
			fmt.Fprintln(w, "Hello_world from device via shifu!")
		}

		func headers(w http.ResponseWriter, req *http.Request) {
			for name, headers := range req.Header {
				for _, header := range headers {
					fmt.Fprintf(w, "%v: %v\n", name, header)
				}
			}
		}

		func main() {
			http.HandleFunc("/hello", process_hello)
			http.HandleFunc("/headers", headers)

			http.ListenAndServe(":11111", nil)
		}
    ```
   and its Dockerfile:
   ```
		package main
        # syntax=docker/dockerfile:1

        FROM golang:1.17-alpine
        WORKDIR /app
        COPY go.mod ./
        RUN go mod download
        COPY *.go ./
        RUN go build -o /helloworld
        EXPOSE 11111
        CMD [ "/helloworld" ]    
   ```
   You can test it locally but it is not covered here.
   
   Build the image

    `docker build --tag <your_docker_account>/helloworld:v0.0.1 .`

   Upload the image to dockerhub

    `docker push <your_docker_account>/helloworld:v0.0.1`

2. ### Prepare the configuration for the edge device
   The basic information of the edge device.
   Assuming all configurations are saved in `<working_dir>/helloworlddevice/configuration`.

   Deployment for the device:
   **helloworld-deployment.yaml**
    ```
    apiVersion: apps/v1
	kind: Deployment
	metadata:
	  labels:
	    app: helloworld
	  name: helloworld
	  namespace: devices
	spec:
	  replicas: 1
	  selector:
	    matchLabels:
	      app: helloworld
	  strategy: {}
	  template:
	    metadata:
	      labels:
	        app: helloworld
	    spec:
	      containers:
	        - image: <your_docker_account>/helloworld:v0.0.1
	          name: helloworld
	          ports:
	            - containerPort: 11111
	          env:
	            - name: MOCKDEVICE_NAME
	              value: helloworld
	            - name: MOCKDEVICE_PORT
	              value: "11111"
      ```
   
   Hardware info for the device:
    **helloworld-edgedevice.yaml**
    ```
	 apiVersion: shifu.edgenesis.io/v1alpha1
	 kind: EdgeDevice
	 metadata:
	   name: edgedevice-helloworld
	   namespace: devices
	 spec:
	   sku: "Hello World"
	   connection: Ethernet
	   address: helloworld.devices.svc.cluster.local:11111
	   protocol: HTTP
	 status:
	   edgedevicephase: "Pending"
	  ```

    Service for the device:
    **helloworld-service.yaml**
   ```
     apiVersion: v1
	 kind: Service
	 metadata:
	   labels:
	     app: helloworld
	   name: helloworld
	   namespace: devices
	 spec:
	   ports:
	     - port: 11111
	       protocol: TCP
	       targetPort: 11111
	   selector:
	     app: helloworld
	   type: LoadBalancer
   ```
3. ### Prepare the configurations for *shifu* to create the *deviceShifu* (digital twin)
   With the following configurations, *shifu* is able to create a *deviceShifu* (digital twin) automatically for the device.
   Assuming all configurations are saved in `<working_dir>/helloworlddevice/configuration`.

   ConfigMap for the deviceShifu:
   **deviceshifu-helloworld-configmap.yaml**
    ```
	apiVersion: v1
	kind: ConfigMap
	metadata:
	  name: helloworld-configmap-0.0.1
	  namespace: default
	data:
	  driverProperties: |
	#    device name and image address
	    driverSku: Hello World
	    driverImage: <your_docker_account>/helloworld:v0.0.1
	  instructions: |
	#    available instructions
	    hello:
	  telemetries: |
	#    telemetry retrieval methods
	#    in this example, a device_health telemetry is collected by calling hello instruction every 1 second
	    device_health:
	      properties:
	        instruction: hello
	        initialDelayMs: 1000
	        intervalMs: 1000
   ```
   Deployment for the deviceShifu:
   **deviceshifu-helloworld-deployment**
    ```
	apiVersion: apps/v1
	kind: Deployment
	metadata:
	  labels:
	    app: edgedevice-helloworld-deployment
	  name: edgedevice-helloworld-deployment
	  namespace: default
	spec:
	  replicas: 1
	  selector:
	    matchLabels:
	      app: edgedevice-helloworld-deployment
	  template:
	    metadata:
	      labels:
	        app: edgedevice-helloworld-deployment
	    spec:
	      containers:
	        - image: edgehub/deviceshifu-http:v0.0.1
	          name: deviceshifu-http
	          ports:
	            - containerPort: 8080
	          volumeMounts:
	            - name: edgedevice-config
	              mountPath: "/etc/edgedevice/config"
	              readOnly: true
	          env:
	            - name: EDGEDEVICE_NAME
	              value: "edgedevice-helloworld"
	            - name: EDGEDEVICE_NAMESPACE
	              value: "devices"
	      volumes:
	        - name: edgedevice-config
	          configMap:
	            name: helloworld-configmap-0.0.1
	      serviceAccountName: edgedevice-readwrite-sa   
   ```
    Service for the deviceShifu:
    **deviceshifu-helloworld-service.yaml**
    ```
	 apiVersion: v1
	 kind: Service
	 metadata:
	   labels:
	     app: edgedevice-helloworld-deployment
	   name: edgedevice-helloworld
	   namespace: default
	 spec:
	   ports:
	     - port: 80
	       protocol: TCP
	       targetPort: 8080
	   selector:
	     app: edgedevice-helloworld-deployment
	   type: LoadBalancer
   ```

4. ### Start *Shifu* and create *deviceShifu*
   Now we have everything ready, and it is the time to start *shifu* and connect the device.
   Assuming the source code of *shifu* is checked out in the working directory (`cd shifu` will go into the *shifu* project root directory).
   1. create Kind cluster (Kubernetes in docker)
       `kind create cluster`
   2. build and load the Kubernetes CRD controller
       ```
       cd shifu/k8s/crd
       make install
       make deploy IMG=edgehub/edgedevice-controller:v0.0.1
       kubectl config set-context --current --namespace=crd-system
       ```
   3. load the Hello World Device docker image
       `kind load docker-image <your_docker_account>/helloworld:v0.0.1`
   4. start Shifu service
       ```
       # return to shifu root directory
       cd ../../shifu
       ./test/scripts/deviceshifu-setup.sh
       ```
   5. let *shifu* create the *deviceShifu* from the configurations
       `kubectl apply -f <working_dir>/helloworlddevice/configuration`
   6. start a nginx server
       `kubectl run nginx --image=nginx:1.21`
      Now we should have the following pods:
        ```
        kubectl get pods --all-namespaces
		NAMESPACE            NAME                                                READY   STATUS    RESTARTS   AGE
		crd-system           crd-controller-manager-7bc78896b9-sq72b             2/2     Running   0          28m
		default              edgedevice-helloworld-deployment-6464b55979-hbdhr   1/1     Running   0          27m
		default              nginx                                               1/1     Running   0          8s
		devices              helloworld-5f467bf5db-f5hxh                         1/1     Running   0          25m
		kube-system          coredns-558bd4d5db-qqx92                            1/1     Running   0          30m
		kube-system          coredns-558bd4d5db-zlw8b                            1/1     Running   0          30m
		kube-system          etcd-kind-control-plane                             1/1     Running   0          30m
		kube-system          kindnet-ndrnh                                       1/1     Running   0          30m
		kube-system          kube-apiserver-kind-control-plane                   1/1     Running   0          30m
		kube-system          kube-controller-manager-kind-control-plane          1/1     Running   0          30m
		kube-system          kube-proxy-qkswm                                    1/1     Running   0          30m
		kube-system          kube-scheduler-kind-control-plane                   1/1     Running   0          30m
		local-path-storage   local-path-provisioner-547f784dff-44xnv             1/1     Running   0          30m
      ```
   7. get into the nginx shell
       `kubectl exec -it --namespace default nginx -- bash`
   8. try interact with the Hellow World Device
      `/# curl http://edgedevice-helloworld:80/hello`

      you should be able to see this:
      `Hello_world from device via shifu!`

Now the Hello World Device is fully integrated in the *shifu* framework and we can interact with it via the *deviceShifu* as shown above.