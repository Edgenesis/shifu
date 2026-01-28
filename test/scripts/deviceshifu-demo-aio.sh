#!/bin/bash
# 2022 Edgenesis Inc.
set -e

if [ $# -eq 0 ]; then
    echo "Usage: $0 <action> [arch] [os]"
    echo "  action: 'run_demo', 'build_demo' or 'delete_demo'"
    echo "  arch: 'amd64' or 'arm64' (optional, defaults to system arch)"
    echo "  os: 'linux' or 'darwin' (optional, defaults to system os)"
    exit 1
fi

SHIFU_IMG_VERSION=v0.89.0-rc1
BUILD_DIR=build_dir
IMG_DIR=images
RUN_DIR=run_dir
UTIL_DIR=util_dir
SHIFU_DIR=shifu
RESOURCE_DIR=resource
BIN_DIR=/usr/local/bin

SHIFU_IMG_LIST=(
    'edgehub/deviceshifu-http-http'
    'edgehub/deviceshifu-http-socket'
    'edgehub/deviceshifu-http-mqtt'
    'edgehub/deviceshifu-http-opcua'
    'edgehub/shifu-controller'
    'edgehub/mockdevice-thermometer'
    'edgehub/mockdevice-robot-arm'
    'edgehub/mockdevice-plate-reader'
    'edgehub/mockdevice-agv'
    'edgehub/mockdevice-plc'
    'edgehub/mockdevice-socket'
    'edgehub/mockdevice-opcua'
)

KIND_IMG="kindest/node:v1.31.0"
KIND_VERSION="v0.24.0"

UTIL_IMG_LIST=(
    $KIND_IMG
    'nginx:1.21'
    'eclipse-mosquitto:2.0.14'
)

# Determine architecture
arch=$(uname -m)
build_arch=""
if [[ ($# -ge 2 && ($2 == "amd64" || $2 == "arm64")) ]]; then
    build_arch=$2
elif [[ $arch == x86_64* ]]; then
    build_arch="amd64"
elif [[ $arch == aarch* ]] || [[ $arch == arm64* ]]; then
    build_arch="arm64"
else
    echo "No support for CPU arch: $arch, exiting..."
    exit 1
fi

# Determine OS
os=$(uname -s)
build_os=""
if [[ ($# -ge 3 && ($3 == "linux" || $3 == "darwin")) ]]; then
    build_os=$3
elif [[ $os == Linux* ]]; then
    build_os="linux"
elif [[ $os == Darwin* ]]; then
    build_os="darwin"
else
    echo "No support for OS: $os, exiting..."
    exit 1
fi

# Validate supported combinations (only linux/amd64, linux/arm64, darwin/arm64)
if [[ $build_os == "darwin" && $build_arch == "amd64" ]]; then
    echo "Error: Darwin amd64 is not supported. Only darwin/arm64 is supported for macOS."
    exit 1
fi

# Supported combinations: linux/amd64, linux/arm64, darwin/arm64
supported_combinations=("linux/amd64" "linux/arm64" "darwin/arm64")
current_combo="$build_os/$build_arch"
if [[ ! " ${supported_combinations[@]} " =~ " ${current_combo} " ]]; then
    echo "Unsupported combination: $current_combo"
    echo "Supported combinations: ${supported_combinations[*]}"
    exit 1
fi

aio_tar_gz_name=shifu_demo_aio_"$build_os"_"$build_arch".tar.gz
aio_tar_name=shifu_demo_aio_"$build_os"_"$build_arch".tar

echo "running on $build_os/$build_arch, tar name should be $aio_tar_name"


if [ $1 = "build_demo" ]; then
    echo "building demo"
    for shifu_image in "${SHIFU_IMG_LIST[@]}"
    do
        echo "docker pull ""$shifu_image":$SHIFU_IMG_VERSION
        docker pull --platform=linux/$build_arch "$shifu_image":$SHIFU_IMG_VERSION
    done

    for util_image in "${UTIL_IMG_LIST[@]}"
    do
        echo "docker pull ""$util_image"
        docker pull --platform=linux/$build_arch "$util_image"
    done

    rm -rf $BUILD_DIR
    mkdir -p $BUILD_DIR/$IMG_DIR

    for shifu_image in "${SHIFU_IMG_LIST[@]}"
    do
        IFS='/' read -r -a array <<< "$shifu_image"
        last_element=${array[${#array[@]}-1]}
        IFS=':' read -r -a array2 <<< "$last_element"
        tar_name=${array2[0]}".tar.gz"
        echo "compressing: "$shifu_image":$SHIFU_IMG_VERSION"
        docker save "$shifu_image":$SHIFU_IMG_VERSION | gzip > $BUILD_DIR/$IMG_DIR/$tar_name
    done

    for util_image in "${UTIL_IMG_LIST[@]}"
    do
        IFS='/' read -r -a array <<< "$util_image"
        last_element=${array[${#array[@]}-1]}
        IFS=':' read -r -a array2 <<< "$last_element"
        tar_name=${array2[0]}".tar.gz"
        echo "compressing: "$util_image""
        docker save "$util_image" | gzip > $BUILD_DIR/$IMG_DIR/$tar_name
    done

    mkdir -p $BUILD_DIR/$UTIL_DIR
    curl -Lo $BUILD_DIR/$UTIL_DIR/kind https://kind.sigs.k8s.io/dl/$KIND_VERSION/kind-$build_os-$build_arch
    curl -Lo $BUILD_DIR/$UTIL_DIR/kubectl "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/$build_os/$build_arch/kubectl"

    mkdir -p $BUILD_DIR/$SHIFU_DIR
    cp pkg/k8s/crd/install/shifu_install.yml $BUILD_DIR/$SHIFU_DIR

    cp -R examples $BUILD_DIR/$SHIFU_DIR
    mv $BUILD_DIR/$SHIFU_DIR/examples/deviceshifu/demo_device $BUILD_DIR/$SHIFU_DIR
    
    echo "compressing gz $aio_tar_gz_name"
    rm -f $aio_tar_gz_name
    tar -czvf $aio_tar_gz_name -C $BUILD_DIR .

    echo "compressing final tar $aio_tar_name"
    rm -f $aio_tar_name
    tar -czvf $aio_tar_name $aio_tar_gz_name test/scripts/deviceshifu-demo-aio.sh

    echo "Finished compressing tar!"
elif [[ $1 = "run_demo" ]]; then
    echo "running demo"
    IP_ADDRESS=$( curl cip.cc --connect-timeout 5 | head -n1 | cut -d ":" -f 2 | xargs)
    curl -X POST https://telemetry.shifu.dev/demo-stat/ \
        -H 'Content-Type: application/json' \
        -d "{\"ip\":\"$IP_ADDRESS\",\"source\":\"shifu_demo_installation_script\",\"task\":\"run_demo_script\",\"step\":\"begin\"}" > /dev/null 2>&1 || true
    rm -rf $RUN_DIR
    mkdir -p $RUN_DIR
    tar -xzvf $aio_tar_gz_name -C $RUN_DIR

    curl -X POST https://telemetry.shifu.dev/demo-stat/ \
        -H 'Content-Type: application/json' \
        -d "{\"ip\":\"$IP_ADDRESS\",\"source\":\"shifu_demo_installation_script\",\"task\":\"run_demo_script\",\"step\":\"after untar\"}" > /dev/null 2>&1 || true

    echo "installing kind, kubectl"
    (cd $RUN_DIR/$UTIL_DIR && chmod +x ./kind && mv ./kind ${BIN_DIR}/kind)
    (cd $RUN_DIR/$UTIL_DIR && chmod +x ./kubectl && mv ./kubectl ${BIN_DIR}/kubectl)
    (cd $RUN_DIR/$UTIL_DIR && ls -lh)

    curl -X POST https://telemetry.shifu.dev/demo-stat/ \
        -H 'Content-Type: application/json' \
        -d "{\"ip\":\"$IP_ADDRESS\",\"source\":\"shifu_demo_installation_script\",\"task\":\"run_demo_script\",\"step\":\"after kind and kubectl install\"}" > /dev/null 2>&1 || true

    (cd $RUN_DIR/$IMG_DIR && for f in *.tar.gz; do docker load < $f; done)

    ${BIN_DIR}/kind delete cluster
    ${BIN_DIR}/kind create cluster --image=$KIND_IMG

    curl -X POST https://telemetry.shifu.dev/demo-stat/ \
        -H 'Content-Type: application/json' \
        -d "{\"ip\":\"$IP_ADDRESS\",\"source\":\"shifu_demo_installation_script\",\"task\":\"run_demo_script\",\"step\":\"after kind cluster create\"}" > /dev/null 2>&1 || true

    for shifu_image in "${SHIFU_IMG_LIST[@]}"
    do
        ${BIN_DIR}/kind load docker-image $shifu_image:$SHIFU_IMG_VERSION
    done

    curl -X POST https://telemetry.shifu.dev/demo-stat/ \
        -H 'Content-Type: application/json' \
        -d "{\"ip\":\"$IP_ADDRESS\",\"source\":\"shifu_demo_installation_script\",\"task\":\"run_demo_script\",\"step\":\"after kind load Shifu images\"}" > /dev/null 2>&1 || true

    for util_image in "${UTIL_IMG_LIST[@]}"
    do
        if [ $util_image = $KIND_IMG ];then
            continue
        fi
        ${BIN_DIR}/kind load docker-image $util_image
    done

    curl -X POST https://telemetry.shifu.dev/demo-stat/ \
        -H 'Content-Type: application/json' \
        -d "{\"ip\":\"$IP_ADDRESS\",\"source\":\"shifu_demo_installation_script\",\"task\":\"run_demo_script\",\"step\":\"after kind load Util images\"}" > /dev/null 2>&1 || true

    ${BIN_DIR}/kubectl apply -f $RUN_DIR/$SHIFU_DIR/shifu_install.yml
    ${BIN_DIR}/kubectl apply -f $RUN_DIR/$SHIFU_DIR/demo_device/edgedevice-agv

    curl -X POST https://telemetry.shifu.dev/demo-stat/ \
        -H 'Content-Type: application/json' \
        -d "{\"ip\":\"$IP_ADDRESS\",\"source\":\"shifu_demo_installation_script\",\"task\":\"run_demo_script\",\"step\":\"after kubectl apply\"}" > /dev/null 2>&1 || true

    echo "Finished setting up Demo !"
elif [[ $1 = "delete_demo" ]]; then
    ${BIN_DIR}/kind delete cluster
    for shifu_image in "${SHIFU_IMG_LIST[@]}"
    do
        docker rmi $shifu_image:$SHIFU_IMG_VERSION
    done

    for shifu_image in "${UTIL_IMG_LIST[@]}"
    do
        docker rmi $shifu_image
    done

    cd .. && rm -rf $TEST_DIR

    rm -rf /usr/local/bin/kind
    rm -rf /usr/local/bin/kubectl
    echo "Delete shifu Success!"
else
        echo "not a valid argument, need to be build_demo/run_demo"
        exit 0
fi
