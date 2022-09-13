/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Note: the example only works with the code within the same release/branch.
package telemetry

import (
	"context"
	"log"
	"time"

	"github.com/edgenesis/shifu/pkg/k8s/crd/telemetry/types"
	"github.com/edgenesis/shifu/pkg/k8s/crd/telemetry/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func StartTelemetry() {
	for {
		publicIP, err := utils.GetPublicIPAddr(utils.URL_EXTERNAL_IP)
		if err != nil {
			log.Printf("issue getting Public IP")
			publicIP = utils.URL_DEFAULT_PUBLIC_IP
		}

		log.Printf("Public IP is %v\n", publicIP)
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		kVersion, err := clientset.ServerVersion()
		if err != nil {
			panic(err.Error())
		}
		log.Printf("%#v", kVersion)
		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		deploy, err := clientset.AppsV1().Deployments("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		podList := make([]string, len(pods.Items))
		deploymentList := make([]string, len(deploy.Items))
		for index, item := range pods.Items {
			podList[index] = item.Name
		}

		for index, item := range deploy.Items {
			deploymentList[index] = item.Name
		}

		clusterInfoTelemetry := types.ClusterInfo{
			NumPods:           len(podList),
			NumDeployments:    len(deploymentList),
			Pods:              podList,
			Deployments:       deploymentList,
			KubernetesVersion: kVersion.GitVersion,
			Platform:          kVersion.Platform,
		}

		controllerTelemetry := types.TelemetryResponse{
			IP:          publicIP,
			Source:      utils.SOURCE_SHIFU_CONTROLLER,
			Task:        utils.TASK_RUN_DEMO_KIND,
			ClusterInfo: clusterInfoTelemetry,
		}

		if result := utils.SendTelemetry(controllerTelemetry); result == nil {
			log.Println("telemetry done")
		}

		time.Sleep(60 * time.Second)
	}
}
