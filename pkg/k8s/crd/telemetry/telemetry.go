package telemetry

import (
	"context"
	"time"

	"github.com/edgenesis/shifu/pkg/k8s/crd/telemetry/types"
	"github.com/edgenesis/shifu/pkg/k8s/crd/telemetry/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

func StartTelemetry() {
	for {
		publicIP, err := utils.GetPublicIPAddr(utils.URL_EXTERNAL_IP)
		if err != nil {
			klog.Errorf("issue getting Public IP")
			publicIP = utils.URL_DEFAULT_PUBLIC_IP
		}

		klog.Infof("Public IP is %v\n", publicIP)
		config, err := rest.InClusterConfig()
		if err != nil {
			klog.Errorln("error when get cluster Config,error: ", err)
			continue
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			klog.Errorln("cannot get ClusterInfo,errors: ", err)
			continue
		}

		kVersion, err := clientset.ServerVersion()
		if err != nil {
			klog.Errorln("cannot get Kubernetes Server Info,errors: ", err)
			continue
		}
		klog.Infof("%#v", kVersion)
		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			klog.Errorln("cannot get Pod Info,errors: ", err)
			continue
		}

		deploy, err := clientset.AppsV1().Deployments("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			klog.Errorln("cannot get Deployment Info,errors: ", err)
			continue
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
			klog.Infoln("telemetry done")
		}

		time.Sleep(time.Duration(utils.TelemetryIntervalInSecond) * time.Second)
	}
}
