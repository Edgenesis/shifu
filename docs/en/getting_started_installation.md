# Getting started: Installation

## Dependencies
Make sure the following dependencies are installed on the machine:
1. [Golang](https://golang.org/dl/): Shifu is created using Golang.
2. [Docker](https://docs.docker.com/get-docker/): Shifu's components are packaged into a Docker image.
3. [kind](https://kubernetes.io/docs/tasks/tools/): Kind is used to run Kubernetes local cluster with a Docker image.
4. [kubectl](https://kubernetes.io/docs/tasks/tools/): The commandline tool for Kubernetes operations.
5. [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder): Kubebuilder is used for building CRDs.

## Quick setup
 Apply the `shifu_install.yml` file (located in ``shifu/k8s/crd/install/shifu_install.yml``) for the quickest installation:

```
cd shifu
kubectl apply -f k8s/crd/install/shifu_install.yml
```

## Step by step setup
A step by step steup is as follows:
```
1. Initialize CRD
cd shifu/k8s/crd
make kube-builder-init

2. Create a new kind cluster
// kind delete cluster (in case you have any active kind clusters)
kind create cluster

3. Install CRD
make install
```
Now a basic Shifu is running on your machine. The next step is to add new devices to it and start managing them.



