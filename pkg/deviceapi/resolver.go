package deviceapi

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
)

const (
	deviceShifuNamespace = "deviceshifu"
	edgeDeviceNameEnv    = "EDGEDEVICE_NAME"
)

// Resolver reads Kubernetes resources and resolves device metadata.
type Resolver struct {
	clientset       kubernetes.Interface
	edgeDeviceLister func(ctx context.Context) ([]v1alpha1.EdgeDevice, error)
}

// NewResolver creates a Resolver using client-go for core resources and
// a custom lister for EdgeDevice CRDs.
func NewResolver(clientset kubernetes.Interface, edgeDeviceLister func(ctx context.Context) ([]v1alpha1.EdgeDevice, error)) *Resolver {
	return &Resolver{
		clientset:       clientset,
		edgeDeviceLister: edgeDeviceLister,
	}
}

// deviceDeploymentInfo holds the resolved info from a DeviceShifu Deployment.
type deviceDeploymentInfo struct {
	ServiceName   string
	ConfigMapName string
}

// resolveDeployment scans DeviceShifu Deployments in the deviceshifu namespace
// for one matching the given device name via the EDGEDEVICE_NAME env var.
func (r *Resolver) resolveDeployment(ctx context.Context, deviceName string) (*deviceDeploymentInfo, error) {
	deployments, err := r.clientset.AppsV1().Deployments(deviceShifuNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing deployments in %s: %w", deviceShifuNamespace, err)
	}

	for i := range deployments.Items {
		dep := &deployments.Items[i]
		if matchesDevice(dep, deviceName) {
			info := &deviceDeploymentInfo{
				ServiceName: dep.Name + "." + deviceShifuNamespace + ".svc.cluster.local",
			}
			info.ConfigMapName = findConfigMapName(dep)
			return info, nil
		}
	}

	return nil, nil
}

// matchesDevice checks if a Deployment has EDGEDEVICE_NAME set to deviceName.
func matchesDevice(dep *appsv1.Deployment, deviceName string) bool {
	for _, container := range dep.Spec.Template.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == edgeDeviceNameEnv && env.Value == deviceName {
				return true
			}
		}
	}
	return false
}

// findConfigMapName extracts the ConfigMap name from the Deployment's volumes.
func findConfigMapName(dep *appsv1.Deployment) string {
	for _, vol := range dep.Spec.Template.Spec.Volumes {
		if vol.ConfigMap != nil {
			return vol.ConfigMap.Name
		}
	}
	return ""
}

// parseInstructions reads the ConfigMap and parses the instructions key.
func (r *Resolver) parseInstructions(ctx context.Context, configMapName string) ([]Interaction, error) {
	if configMapName == "" {
		return nil, nil
	}

	cm, err := r.clientset.CoreV1().ConfigMaps(deviceShifuNamespace).Get(ctx, configMapName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting configmap %s/%s: %w", deviceShifuNamespace, configMapName, err)
	}

	return parseInstructionsFromConfigMap(cm)
}

